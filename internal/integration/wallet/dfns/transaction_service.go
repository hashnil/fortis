package dfns

import (
	"errors"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/pkg/utils"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// InitTransferAssets provides asset transfers payload including fees for user signing
func (p *DFNSWalletProvider) InitTransferAssets(request models.InitTransferRequest) (*models.InitTransferResponse, error) {
	// Validate user and fetch login token
	sender, loginInfo, err := p.getUserAndLoginInfo(request.UserID)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] User and login info fetched for user: %s", sender.Name)

	// Retrieve Wallets - Sender & Receiver
	primaryNetwork := viper.GetString("wallet.dfns.asset_transfer.primary_network")
	senderWallet, err := p.dbClient.FindWalletByNameAndNetwork(sender.Name, constants.DFNS, primaryNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sender wallet %s: %w", sender.Name, err)
	}
	recipientWallet, err := p.dbClient.FindWalletByNameAndNetwork(request.ToAccount, constants.DFNS, primaryNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve receiver wallet %s: %w", request.ToAccount, err)
	}

	// Process Fund Transfer
	inflightFundTxn, err := p.handleTransactionChallenge(request, senderWallet, recipientWallet.Address, request.Amount, loginInfo.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to process fund transfer: %w", err)
	}

	// Process Fee Transfer
	feeRecipient := viper.GetString("wallet.fees.address")
	inflightFeeTxn, err := p.handleTransactionChallenge(request, senderWallet, feeRecipient, request.Fee, loginInfo.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to process fee transfer: %w", err)
	}

	log.Printf("[INFO] Successfully initialized asset transfer for user: %s", request.UserID)
	return &models.InitTransferResponse{
		Result: constants.SUCCESS,
		Challenge: map[string]string{
			constants.FundTransferChallenge: inflightFundTxn.Challenge,
			constants.FeeTransferChallenge:  inflightFeeTxn.Challenge,
		},
	}, nil
}

// getUserAndLoginInfo retrieves user details and login token
func (p *DFNSWalletProvider) getUserAndLoginInfo(userID string) (*dbmodels.User, *models.LoginToken, error) {
	sender, err := p.dbClient.FindUserByID(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve user %s from DB: %w", userID, err)
	}
	if !sender.IsActive {
		return nil, nil, errors.New(constants.InactiveUser + userID)
	}

	loginInfo, err := APIClient[models.LoginToken](map[string]string{"username": sender.Name}, "POST", constants.DelegatedLoginURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get authentication token: %w", err)
	}
	return &sender, loginInfo, nil
}

// createTransferRequest generates a transfer request
func createTransferRequest(amount string, denom string, recipient string) (models.DFNSTransactionRequest, error) {
	smallestAmount, err := utils.ConvertToSmallestUnit(amount, strings.ToUpper(denom))
	if err != nil {
		return models.DFNSTransactionRequest{}, fmt.Errorf("failed to convert amount to smallest unit: %w", err)
	}

	// Initialize the base transaction request with essential details.
	transferRequest := models.DFNSTransactionRequest{
		Kind:   viper.GetString("wallet.dfns.asset_transfer.native_token"),
		To:     recipient,
		Amount: smallestAmount,
	}

	// Determine the primary blockchain network and configure additional fields accordingly.
	primaryNetwork := viper.GetString("wallet.dfns.asset_transfer.primary_network")

	if primaryNetwork == constants.Solana {
		// Assign mint address and enable ATA (Associated Token Account) creation for Solana transactions.
		transferRequest.Mint = viper.GetString(fmt.Sprintf("wallet.dfns.mint_address.%s", strings.ToLower(denom)))
		transferRequest.CreateATA = true
	} else if primaryNetwork == constants.Base {
		// Assign contract address for Base chain transactions. (TODO: test)
		transferRequest.Contract = viper.GetString(fmt.Sprintf("wallet.dfns.mint_address.%s", strings.ToLower(denom)))
	}

	return transferRequest, nil
}

// handleTransactionChallenge manages the challenge process
func (p *DFNSWalletProvider) handleTransactionChallenge(
	request models.InitTransferRequest, senderWallet dbmodels.Wallet, recipient, amount, authToken string,
) (*dbmodels.InflightTransaction, error) {
	log.Printf("[INFO] Creating transfer request for %s %s to %s", amount, request.Denom, recipient)

	// Extract sender wallet ID
	var senderWalletInfo models.DFNSWalletResponse
	utils.UnmarshalFromJSON(senderWallet.Metadata, &senderWalletInfo)
	senderWalletID := senderWalletInfo.ID

	// Create transaction request
	txRequest, err := createTransferRequest(amount, request.Denom, recipient)
	if err != nil {
		return nil, err
	}

	// Create user challenge payload
	txUserChallengePayload := models.UserActionSignatureChallengeRequest{
		UserActionPayload:    string(utils.MarshalToJSON(txRequest)),
		UserActionHTTPMethod: "POST",
		UserActionHTTPPath:   fmt.Sprintf(constants.TransferAssetsURL, senderWalletID),
	}

	// Call user action signature challenge API
	txChallengeResponse, err := APIClient[models.UserActionSignatureChallengeResponse](txUserChallengePayload, "POST", constants.UserActionSignatureChallengeURL, &authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate user action signature challenge: %w", err)
	}

	// Save transaction payload
	inflightTxn := &dbmodels.InflightTransaction{
		Challenge:            txChallengeResponse.Challenge,
		ChallengeIdentifier:  txChallengeResponse.ChallengeIdentifier,
		URL:                  txUserChallengePayload.UserActionHTTPPath,
		AuthToken:            authToken,
		RequestPayload:       utils.MarshalToJSON(request),
		TransferPayload:      utils.MarshalToJSON(txRequest),
		UserChallengePayload: utils.MarshalToJSON(txChallengeResponse),
		SenderInfo:           utils.MarshalToJSON(senderWallet),
	}

	if err := p.dbClient.CreateInflightTransaction(*inflightTxn); err != nil {
		return nil, fmt.Errorf("failed to save transaction in DB: %w", err)
	}
	log.Printf("[INFO] Transaction challenge saved successfully")

	return inflightTxn, nil
}

// TODO: refactor
func (p *DFNSWalletProvider) TransferAssets(request models.TransferRequest) (*models.TransferResponse, error) {
	response := models.TransferResponse{Result: constants.SUCCESS}
	UTR := utils.GenerateUTR()

	for challenge, _ := range request.CredentialInfo {
		// Fetch the inflight transaction details from DB
		inflightTx, err := p.dbClient.GetInflightTransaction(challenge)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch inflight transaction for challenge %s: %v", challenge, err)
		}

		// TODO: Prepare user action signature payload based on credential type
		// userActionPayload := models.UserActionSigningRequest{
		// 	ChallengeIdentifier: inflightTx.ChallengeIdentifier,
		// }

		// // TODO: Prepare user action signature payload
		// userActionPayload = models.UserActionSigningRequest{
		// 	ChallengeIdentifier: inflightTx.ChallengeIdentifier,
		// 	FirstFactor: models.FirstFactor{
		// 		Kind: credentials.CredentialKind,
		// 		CredentialAssertion: models.CredentialAssertion{
		// 			CredID:            credentials.CredentialID,
		// 			ClientData:        credentials.ClientData,
		// 			AuthenticatorData: credentials.AttestationData,
		// 			Signature:         credentials.Signature,
		// 			UserHandle:        credentials.UserHandle,
		// 		},
		// 	},
		// }

		// Call user action signature API
		// _, err = APIClient[models.UserActionSigningResponse](
		// 	userActionPayload, "POST", constants.UserActionSignatureChallengeURL, &inflightTx.AuthToken,
		// )
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to initiate user action signature challenge for %s: %v", challenge, err)
		// }

		// Unmarshal data
		var (
			initTransferReq models.InitTransferRequest
			transferRequest models.DFNSTransactionRequest
			senderWallet    dbmodels.Wallet
		)
		utils.UnmarshalFromJSON(inflightTx.RequestPayload, &initTransferReq)
		utils.UnmarshalFromJSON(inflightTx.TransferPayload, &transferRequest)
		utils.UnmarshalFromJSON(inflightTx.SenderInfo, &senderWallet)

		// Attempt transaction broadcast
		txResponse, err := APIClient[models.DFNSTransactionResponse](transferRequest, "POST", inflightTx.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast transaction for %s: %w", inflightTx.URL, err)
		}

		// Identify if this is a fee transaction
		typeFee, amount := false, initTransferReq.Amount
		if txResponse.RequestBody.To == viper.GetString("wallet.fees.address") {
			typeFee, amount = true, initTransferReq.Fee
		}

		// Populate response for non-fee transactions
		if !typeFee {
			response.ReceiverID = initTransferReq.ToAccount
			response.Amount = initTransferReq.Amount
			response.Fee = initTransferReq.Fee
			response.Denom = initTransferReq.Denom
			response.UTR = UTR // TODO: remove
			response.TxInfo.ReceiverAddress = txResponse.RequestBody.To
			response.TxInfo.Network = txResponse.Network
			response.TxInfo.TxHash = txResponse.TxHash
		}

		// Create transaction log entry and store in DB
		transactionLog := dbmodels.TransactionLog{
			ID:              fmt.Sprintf("tx-%d", time.Now().UnixNano()),
			SenderName:      senderWallet.Name,
			SenderAddress:   senderWallet.Address,
			ReceiverName:    initTransferReq.ToAccount,
			ReceiverAddress: txResponse.RequestBody.To,
			Amount:          amount,
			Denom:           initTransferReq.Denom,
			Provider:        constants.DFNS,
			Network:         txResponse.Network,
			TypeFee:         typeFee,
			Status:          txResponse.Status, // Broadcasted
			TxHash:          txResponse.TxHash,
			UTR:             UTR,
			TxMeta:          utils.MarshalToJSON(txResponse),
		}
		if err := p.dbClient.CreateTransactionLog(transactionLog); err != nil {
			return nil, fmt.Errorf("failed to log transaction: %w", err)
		}

		// Insert into transactions table for webhook confirmation
		transactionEntry := dbmodels.Transaction{
			TxHash:          txResponse.TxHash,
			SenderName:      senderWallet.Name,
			SenderAddress:   senderWallet.Address,
			ReceiverName:    initTransferReq.ToAccount,
			ReceiverAddress: txResponse.RequestBody.To,
			Amount:          amount,
			Denom:           initTransferReq.Denom,
			Provider:        constants.DFNS,
			Network:         txResponse.Network,
			TypeFee:         typeFee,
			Status:          txResponse.Status, // Broadcasted
			UTR:             UTR,
		}
		if err := p.dbClient.CreateTransaction(transactionEntry); err != nil {
			return nil, fmt.Errorf("failed to insert transaction record: %w", err)
		}

		// Delete inflight transaction after successful broadcast
		if err := p.dbClient.DeleteInflightTransaction(challenge); err != nil {
			return nil, fmt.Errorf("failed to delete inflight transaction: %w", err)
		}
	}
	return &response, nil
}
