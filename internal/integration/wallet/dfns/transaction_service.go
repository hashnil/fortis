package dfns

import (
	"encoding/json"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	dbmodels "fortis/internal/integration/db/models"
	"math/rand"
	"time"

	"github.com/spf13/viper"
)

func generateUTR() string {
	return fmt.Sprintf("UTR-%d", rand.Int63())
}

func (p *DFNSWalletProvider) TransferAssets(request *models.TransactionRequest) (*models.TransactionResponse, error) {
	// get sender wallet details
	senderWallet, err := p.dbClient.FindUserWallet(request.FromAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender wallet %s details: %w", request.FromAccount, err)
	}
	if senderWallet.ID == "" {
		return nil, fmt.Errorf("wallet for user Sender:%s does not exists", request.FromAccount)
	}

	// get recipient wallet details
	recipientWallet, err := p.dbClient.FindUserWallet(request.ToAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver wallet %s details: %w", request.ToAccount, err)
	}
	if recipientWallet.ID == "" {
		return nil, fmt.Errorf("wallet for user Receiver:%s does not exists", request.ToAccount)
	}

	transactionLog := dbmodels.TransactionLog{
		ID:              fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		SenderName:      request.FromAccount,
		SenderAddress:   senderWallet.Address,
		ReceiverName:    request.ToAccount,
		ReceiverAddress: recipientWallet.Address,
		Amount:          request.Amount,
		Denom:           request.Denom,
		Provider:        constants.DFNS,
		Network:         viper.GetString("wallet.dfns.asset_transfer.primary_network"),
		Status:          "Pending",
		FeeType:         false,
		UTR:             generateUTR(),
	}
	if err := p.dbClient.CreateTransactionLog(transactionLog); err != nil {
		return nil, fmt.Errorf("failed to create transaction log before broadcasting the transaction: %w", err)
	}

	// get dfns wallet info
	var dfnsWalletInfo models.DFNSWalletResponse
	if err := json.Unmarshal(senderWallet.WalletMeta, &dfnsWalletInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal")
	}

	txRequest := models.DFNSTransactionRequest{
		Kind:       "Spl",
		Mint:       senderWallet.Address,
		To:         recipientWallet.Address,
		Amount:     request.Amount,
		ExternalID: transactionLog.UTR,
	}

	// Broadcast the transaction
	txResponse, err := APIClient[models.DFNSTransactionResponse](txRequest, "POST", fmt.Sprintf("/wallets/%s/transfers", dfnsWalletInfo.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	fmt.Println("txResponse: ", txResponse)

	return nil, nil
}
