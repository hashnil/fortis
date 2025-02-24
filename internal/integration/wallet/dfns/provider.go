package dfns

import (
	"fmt"
	"fortis/entity/models"
	"fortis/internal/integration/db"
	"fortis/internal/integration/wallet"
)

type DFNSWalletProvider struct {
	dbClient db.Client
}

func NewDFNSWalletProvider(dbClient db.Client) wallet.Provider {
	provider := &DFNSWalletProvider{
		dbClient: dbClient,
	}
	provider.registerWebhook()
	return provider
}

// registerWebhook registers a webhook for DFNS wallet transfer events.
func (p *DFNSWalletProvider) registerWebhook() error {
	webhookRequest := models.DFNSWebhookRequest{
		URL:         p.WebhookURL,
		Description: "Webhook subscribing to transactional events",
		Events: []string{
			"wallet.transfer.requested",
			"wallet.transfer.failed",
			"wallet.transfer.rejected",
			"wallet.transfer.broadcasted",
			"wallet.transfer.confirmed",
		},
	}

	webhookResponse, err := APIClient[models.DFNSWebhookResponse](webhookRequest, "POST", "/wallets")
	if err != nil {
		return fmt.Errorf("failed to create webhook for events %w: %w", webhookRequest, err)
	}
	fmt.Println("webhookResponse: ", webhookResponse)

	return nil
}
