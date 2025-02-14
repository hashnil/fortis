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
	dbClient            db.Client       // Database client for user operations.
	firstWalletProvider wallet.Provider // Wallet provider for wallet operations: DFNS
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
		dbClient:            dbClient,
		firstWalletProvider: dfnsWalletProvider,
	}, nil
}

func (c *WalletController) CreateWalletV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.CreateWalletsHandlerV1).Inc()

	var requestBody models.WalletRequest
	if !utils.BindRequest(ctx, &requestBody, constants.CreateWalletsHandlerV1, startTime) {
		return
	}

	var walletProvider wallet.Provider
	switch ctx.Param("provider") {
	case constants.DFNS:
		walletProvider = c.firstWalletProvider
	default:
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, nil, constants.CreateWalletsHandlerV1, startTime)
		return
	}

	// Create wallet for the user
	_, err := walletProvider.CreateWallet(&requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateWallet, constants.ErrCreateWallet, err, constants.CreateWalletsHandlerV1, startTime)
		return
	}

	// Return aggregated results
	log.Println("Wallets successfully created for user: " + requestBody.Username)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.CreateWalletsHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.CreateWalletsHandlerV1).Observe(time.Since(startTime).Seconds())
	ctx.JSON(http.StatusOK, models.WalletResponse{Result: constants.SUCCESS})
}
