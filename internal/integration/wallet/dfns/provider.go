package dfns

import (
	"fmt"
	"fortis/entity/models"
	"fortis/internal/integration/db"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/internal/integration/wallet"
	"fortis/pkg/utils"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type DFNSWalletProvider struct {
	dbClient db.Client
}

func NewDFNSWalletProvider(dbClient db.Client) wallet.Provider {
	provider := &DFNSWalletProvider{
		dbClient: dbClient,
	}
	if err := provider.registerWebhook(); err != nil {
		log.Fatalf("failed to register webhook: %v", err)
	}
	go provider.startWebhookListener()
	return provider
}

// TODO: refactor
// registerWebhook registers a webhook for DFNS wallet transfer events.
func (p *DFNSWalletProvider) registerWebhook() error {
	webhookRequest := models.DFNSWebhookRequest{
		URL:         viper.GetString("wallet.dfns.webhook.url"),
		Description: "Webhook subscribing to transactional events",
		Events: []string{
			"wallet.transfer.failed",
			"wallet.transfer.rejected",
			"wallet.transfer.confirmed",
		},
	}

	webhookResponse, err := APIClient[models.DFNSWebhookResponse](webhookRequest, "POST", "/webhooks", nil)
	if err != nil {
		return fmt.Errorf("failed to create webhook for events %v: %w", webhookRequest, err)
	}
	fmt.Println("webhookResponse: ", webhookResponse)

	return nil
}

// startWebhookListener starts an HTTP server to listen for webhook events.
func (p *DFNSWalletProvider) startWebhookListener() {
	http.HandleFunc("http://localhost:8081/webhook", p.handleWebhookEvent)
	fmt.Println("Starting webhook listener on port 8081...")
	http.ListenAndServe(":8081", nil)
}

// handleWebhookEvent processes incoming webhook notifications.
func (p *DFNSWalletProvider) handleWebhookEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event models.DFNSWebhookTransferResponse
	utils.UnmarshalFromJSON(body, &event)

	// Log and process the event based on its type
	fmt.Println("Received DFNS Webhook Event:", event.Kind)
	p.processWebhookEvent(event)

	w.WriteHeader(http.StatusOK)
}

// processWebhookEvent handles specific webhook events.
func (p *DFNSWalletProvider) processWebhookEvent(event models.DFNSWebhookTransferResponse) {
	switch event.Kind {
	case "wallet.transfer.failed":
		fmt.Println("Processing transfer failed event", event)
	case "wallet.transfer.rejected":
		fmt.Println("Processing transfer rejected event", event)
	case "wallet.transfer.confirmed":
		fmt.Println("Processing transfer confirmed event", event)
	default:
		fmt.Println("Unknown event type received", event.Kind)
	}

	// Get non-deleted transaction via hash
	transaction, err := p.dbClient.GetTransaction(event.Data.TransferRequest.TxHash)
	if err != nil {
		fmt.Printf("Error retrieving transaction (TxHash: %s): %v\n", event.Data.TransferRequest.TxHash, err)
		return
	}

	// Store the transaction in transaction log
	logEntry := dbmodels.TransactionLog{
		ID:              uuid.New().String(),
		SenderName:      transaction.SenderName,
		SenderAddress:   transaction.SenderAddress,
		ReceiverName:    transaction.ReceiverName,
		ReceiverAddress: transaction.ReceiverAddress,
		Amount:          transaction.Amount,
		Denom:           transaction.Denom,
		Provider:        transaction.Provider,
		Network:         transaction.Network,
		TypeFee:         transaction.TypeFee,
		Status:          event.Data.TransferRequest.Status,
		FailureReason:   event.Error,
		TxHash:          transaction.TxHash,
		UTR:             transaction.UTR,
		TxMeta:          utils.MarshalToJSON(event),
	}

	if err := p.dbClient.CreateTransactionLog(logEntry); err != nil {
		fmt.Printf("Error storing transaction log: %v\n", err)
		return
	}

	// Delete the transaction entry
	if err := p.dbClient.DeleteTransaction(transaction.TxHash); err != nil {
		fmt.Printf("Error deleting transaction (TxHash: %s): %v\n", transaction.TxHash, err)
	}
}
