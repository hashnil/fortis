package app

import (
	"fmt"
	"fortis/api/controller"
	"fortis/entity/constants"
)

// registerRoutes sets up the routes for the application.
func registerRoutes(healthController *controller.HealthController, walletController *controller.WalletController) {
	// Health check route
	engine.GET("/health", healthController.HealthCheck)

	// API v1 routes grouped under `/api/v1`
	apiV1 := engine.Group(fmt.Sprintf("/api/%s", constants.API_VERSION_V1))

	// Wallet access routes
	apiV1.POST("/:provider/wallet/register-user", walletController.CreateDelegatedUserV1)
	apiV1.POST("/:provider/wallet/create-wallet", walletController.CreateWalletV1)
	apiV1.POST("/:provider/transfer-assets", walletController.TransferAssetsV1)
}
