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
	"strings"
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

// getWalletProvider retrieves the appropriate wallet provider based on the request.
func (c *WalletController) getWalletProvider(ctx *gin.Context) (wallet.Provider, error) {
	switch ctx.Param("provider") {
	case constants.DFNS:
		return c.firstWalletProvider, nil
	default:
		return nil, fmt.Errorf("provider not found: %s", ctx.Param("provider"))
	}
}

func (c *WalletController) CreateDelegatedUserV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.RegisterUserHandlerV1).Inc()

	var requestBody models.CreateUserRequest
	if !utils.BindRequest(ctx, &requestBody, constants.RegisterUserHandlerV1, startTime) {
		return
	}

	// Determine wallet provider
	walletProvider, err := c.getWalletProvider(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, err, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// Create delegated user
	userResponse, err := walletProvider.RegisterDelegatedUser(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrCreateUser, constants.ErrCreateUser, err, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// Handle existing user scenario
	if userResponse.ExistingUser {
		utils.HandleError(ctx, http.StatusConflict,
			constants.ErrExistingUser, constants.ErrExistingUser, err, constants.RegisterUserHandlerV1, startTime)
		return
	}

	// Log success and respond
	log.Printf("Delegated user created successfully: %s", requestBody.Username)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.RegisterUserHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.RegisterUserHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, models.CreateUserResponse{
		Result:    constants.SUCCESS,
		Challenge: userResponse.Challenge})
}

func (c *WalletController) ActivateUserV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.ActivateUserHandlerV1).Inc()

	var requestBody models.ActivateUserRequest
	if !utils.BindRequest(ctx, &requestBody, constants.ActivateUserHandlerV1, startTime) {
		return
	}

	// Determine wallet provider
	walletProvider, err := c.getWalletProvider(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, err, constants.ActivateUserHandlerV1, startTime)
		return
	}

	// TODO: from headers
	requestBody.UserID = constants.UserPrefix + requestBody.UserID

	// Activate user
	err = walletProvider.ActivateDelegatedUser(requestBody)
	if err != nil {
		status := http.StatusInternalServerError
		errMsg := constants.ErrActivateUser
		if strings.Contains(err.Error(), constants.DuplicateUser) {
			status = http.StatusConflict
			errMsg = constants.ErrExistingUser
		}
		utils.HandleError(ctx, status, errMsg, errMsg, err, constants.ActivateUserHandlerV1, startTime)
		return
	}

	// Log success and respond
	log.Printf("User successfully activated: %s", requestBody.UserID)
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

	// Determine wallet provider
	walletProvider, err := c.getWalletProvider(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, err, constants.CreateWalletHandlerV1, startTime)
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
	log.Printf("Wallet successfully created for user: %s", requestBody.UserID)
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

	var requestBody models.TransferRequest
	if !utils.BindRequest(ctx, &requestBody, constants.TransferAssetsHandlerV1, startTime) {
		return
	}

	// Determine wallet provider
	walletProvider, err := c.getWalletProvider(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, err, constants.TransferAssetsHandlerV1, startTime)
		return
	}

	// Transfer assets
	_, err = walletProvider.TransferAssets(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrTransferAssets, constants.ErrTransferAssets, err, constants.TransferAssetsHandlerV1, startTime)
		return
	}

	// Return aggregated results
	log.Println("Assets successfully transfered for user: ")
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.TransferAssetsHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.TransferAssetsHandlerV1).Observe(time.Since(startTime).Seconds())
	ctx.JSON(http.StatusOK, models.TransferResponse{Result: constants.SUCCESS})
}
