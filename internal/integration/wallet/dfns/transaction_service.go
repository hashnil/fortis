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

	return models.DFNSTransactionRequest{
		Kind:   viper.GetString("wallet.dfns.asset_transfer.native_token"),
		Mint:   viper.GetString(fmt.Sprintf("wallet.dfns.mint_address.%s", strings.ToLower(denom))),
		To:     recipient,
		Amount: smallestAmount,
	}, nil
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

func (p *DFNSWalletProvider) TransferAssets(request models.TransferRequest) (*models.TransferResponse, error) {
	// Basis challenge get the inflight transaction from database
	for challenge, credentials := range request.CredentialInfo {
		inflightTx, err := p.dbClient.GetInflightTransaction(challenge)
		if err != nil {
			return nil, err
		}

		userActionPayload := models.UserActionSigningRequest{
			ChallengeIdentifier: inflightTx.ChallengeIdentifier,
			FirstFactor: struct {
				Kind                string "json:\"kind\""
				CredentialAssertion struct {
					CredID            string "json:\"credId\""
					ClientData        string "json:\"clientData\""
					AuthenticatorData string "json:\"authenticatorData\""
					Signature         string "json:\"signature\""
					UserHandle        string "json:\"userHandle\""
				} "json:\"credentialAssertion\""
			}{
				Kind: credentials.CredentialKind,
				CredentialAssertion: struct {
					CredID            string "json:\"credId\""
					ClientData        string "json:\"clientData\""
					AuthenticatorData string "json:\"authenticatorData\""
					Signature         string "json:\"signature\""
					UserHandle        string "json:\"userHandle\""
				}{
					CredID:            credentials.CredentialID,
					ClientData:        credentials.ClientData,
					AuthenticatorData: credentials.AttestationData,
					Signature:         credentials.Signature,
					UserHandle:        credentials.UserHandle,
				},
			},
		}

		// Call user action signature API
		userActionResponse, err := APIClient[models.UserActionSigningResponse](userActionPayload, "POST", constants.UserActionSignatureChallengeURL, &inflightTx.AuthToken)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate user action signature challenge: %w", err)
		}

		var (
			transferRequest models.DFNSTransactionRequest
			request         models.InitTransferRequest
			senderWallet    dbmodels.Wallet
		)
		utils.UnmarshalFromJSON(inflightTx.RequestPayload, &request)
		utils.UnmarshalFromJSON(inflightTx.TransferPayload, &transferRequest)
		utils.UnmarshalFromJSON(inflightTx.SenderInfo, &senderWallet)

		// Call transfer assets API
		// Broadcast the transaction
		txResponse, err := APIClient[models.DFNSTransactionResponse](transferRequest, "POST", inflightTx.URL, &userActionResponse.UserAction)
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
		}

		transactionLog := dbmodels.TransactionLog{
			ID:              fmt.Sprintf("tx-%d", time.Now().UnixNano()),
			SenderName:      senderWallet.Name,
			SenderAddress:   senderWallet.Address,
			ReceiverName:    request.ToAccount,
			ReceiverAddress: txResponse.RequestBody.To,
			Amount:          request.Amount,
			Denom:           request.Denom,
			Provider:        constants.DFNS,
			Network:         txResponse.Network,
			TypeFee:         txResponse.Fee != "",
			Status:          txResponse.Status,
			TxHash:          txResponse.TxHash,
			TxMeta:          utils.MarshalToJSON(txResponse),
			UTR:             utils.GenerateUTR(),
		}
		if err := p.dbClient.CreateTransactionLog(transactionLog); err != nil {
			return nil, fmt.Errorf("failed to create transaction log before broadcasting the transaction: %w", err)
		}

		fmt.Println("txResponse: ", txResponse)
	}
	return nil, nil
}
