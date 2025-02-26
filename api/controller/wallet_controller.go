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
	firstWalletProvider wallet.Provider // Wallet provider for wallet operations: DFNS
}

func NewWalletController(dbClient db.Client) (*WalletController, error) {
	// Initialize DFNS client for wallet operations.
	dfnsWalletProvider, err := factory.InitWalletProvider(constants.DFNS, dbClient)
	if err != nil {
		return &WalletController{}, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	return &WalletController{
		firstWalletProvider: dfnsWalletProvider,
	}, nil
}

func (c *WalletController) CreateDelegatedUserV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.RegisterUserHandlerV1).Inc()

	var requestBody models.CreateUserRequest
	if !utils.BindRequest(ctx, &requestBody, constants.RegisterUserHandlerV1, startTime) {
		return
	}

	var walletProvider wallet.Provider
	switch ctx.Param("provider") {
	case constants.DFNS:
		walletProvider = c.firstWalletProvider
	default:
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, nil, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// Create delegated user
	userResponse, err := walletProvider.RegisterDelegatedUser(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateUser, constants.ErrCreateUser, err, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// User already present
	if userResponse.ExistingUser {
		utils.HandleError(ctx, http.StatusConflict,
			constants.ErrExistingUser, constants.ErrExistingUser, err, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// Return aggregated results
	log.Println("Delegated user created successfully for: " + requestBody.Username)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.RegisterUserHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.RegisterUserHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, models.CreateUserResponse{
		Result:              constants.SUCCESS,
		Challenge:           userResponse.Challenge,
		AuthenticationToken: userResponse.AuthenticationToken})
}

func (c *WalletController) ActivateUserV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.ActivateUserHandlerV1).Inc()

	var requestBody models.ActivateUserRequest
	if !utils.BindRequest(ctx, &requestBody, constants.ActivateUserHandlerV1, startTime) {
		return
	}

	var walletProvider wallet.Provider
	switch ctx.Param("provider") {
	case constants.DFNS:
		walletProvider = c.firstWalletProvider
	default:
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, nil, constants.ActivateUserHandlerV1, startTime)
		return
	}

	// TODO: from headers
	requestBody.UserID = constants.UserPrefix + requestBody.UserID

	// Create wallet for the user
	_, err := walletProvider.ActivateDelegatedUser(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateWallet, constants.ErrCreateWallet, err, constants.ActivateUserHandlerV1, startTime)
		return
	}

	// Return aggregated results
	log.Println("User successfully activated: " + requestBody.UserID)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.ActivateUserHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.ActivateUserHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, models.ActivateUserResponse{Result: constants.SUCCESS})
}

func (c *WalletController) CreateWalletV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.CreateWalletHandlerV1).Inc()

	var requestBody models.WalletRequest
	if !utils.BindRequest(ctx, &requestBody, constants.CreateWalletHandlerV1, startTime) {
		return
	}

	var walletProvider wallet.Provider
	switch ctx.Param("provider") {
	case constants.DFNS:
		walletProvider = c.firstWalletProvider
	default:
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, nil, constants.CreateWalletHandlerV1, startTime)
		return
	}

	// Create wallet for the user
	walletResponse, err := walletProvider.CreateWallet(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateWallet, constants.ErrCreateWallet, err, constants.CreateWalletHandlerV1, startTime)
		return
	}

	// Successful wallet creation
	log.Println("Wallet created successfully for user: " + requestBody.UserID)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.CreateWalletHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.CreateWalletHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, models.WalletResponse{
		Result:    constants.SUCCESS,
		Addresses: walletResponse.Addresses,
	})
}

func (c *WalletController) TransferAssetsV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.TransferAssetsHandlerV1).Inc()

	var requestBody models.TransactionRequest
	if !utils.BindRequest(ctx, &requestBody, constants.TransferAssetsHandlerV1, startTime) {
		return
	}

	var walletProvider wallet.Provider
	switch ctx.Param("provider") {
	case constants.DFNS:
		walletProvider = c.firstWalletProvider
	default:
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, nil, constants.TransferAssetsHandlerV1, startTime)
		return
	}

	// Transfer assets
	_, err := walletProvider.TransferAssets(&requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateWallet, constants.ErrCreateWallet, err, constants.TransferAssetsHandlerV1, startTime)
		return
	}

	// Return aggregated results
	log.Println("Assets successfully transfered for user: " + requestBody.FromAccount)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.TransferAssetsHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.TransferAssetsHandlerV1).Observe(time.Since(startTime).Seconds())
	ctx.JSON(http.StatusOK, models.TransactionResponse{Result: constants.SUCCESS})
}
