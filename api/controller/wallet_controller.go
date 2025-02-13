package controller

import (
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	"fortis/infrastructure/factory"
	"fortis/internal/instrumentation"
	"fortis/internal/integration/db"
	"fortis/internal/integration/wallet"
	"fortis/pkg/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	dbClient       db.Client // Database client for user operations.
	walletProvider wallet.Provider
}

func NewWalletController() (*WalletController, error) {
	// Initialize PostgreSQL client for database operations.
	dbClient, err := factory.InitDBClient(constants.PostgreSQL)
	if err != nil {
		return &WalletController{}, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	// Initialize DFNS client for wallet operations.
	dfnsWalletProvider, err := factory.InitWalletProvider(constants.DFNS, dbClient)
	if err != nil {
		return &WalletController{}, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	return &WalletController{
		dbClient:       dbClient,
		walletProvider: dfnsWalletProvider,
	}, nil
}

func (c *WalletController) CreateWallet(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.CreateWalletsHandler).Inc()

	var requestBody models.WalletRequest
	if !utils.BindRequest(ctx, &requestBody, constants.CreateWalletsHandler, startTime) {
		return
	}

	// Create wallet for the user
	_, err := c.walletProvider.CreateWallet(&requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateWallet, constants.ErrCreateWallet, err, constants.CreateWalletsHandler, startTime)
		return
	}

	// Return aggregated results
	log.Println("Wallets successfully created for user: " + requestBody.Username)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.CreateWalletsHandler).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.CreateWalletsHandler).Observe(time.Since(startTime).Seconds())
	ctx.JSON(http.StatusOK, models.WalletResponse{Result: constants.SUCCESS})
}
