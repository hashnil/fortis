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
	log.Printf("[CreateDelegatedUserV1] Delegated user created successfully: %s\n", requestBody.Username)
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
	log.Printf("[ActivateUserV1] User successfully activated: %s\n", requestBody.UserID)
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
		status := http.StatusInternalServerError
		errMsg := constants.ErrCreateWallet
		if strings.Contains(err.Error(), constants.InactiveUser) {
			status = http.StatusNotFound
			errMsg = constants.ErrInactiveUser
		}
		utils.HandleError(ctx, status, errMsg, errMsg, err, constants.CreateWalletHandlerV1, startTime)
		return
	}

	// Successful wallet creation
	log.Printf("[CreateWalletV1] Wallets %+v successfully created for user: %s\n", walletResponse.Addresses, requestBody.UserID)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.CreateWalletHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.CreateWalletHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, models.WalletResponse{
		Result:    constants.SUCCESS,
		Addresses: walletResponse.Addresses,
	})
}

func (c *WalletController) InitTransferAssetsV1(ctx *gin.Context) {
	// Record API counter and start time for instrumentation.
	startTime := time.Now()
	instrumentation.RequestCounter.WithLabelValues(constants.InitTransferAssetsHandlerV1).Inc()

	var requestBody models.InitTransferRequest
	if !utils.BindRequest(ctx, &requestBody, constants.InitTransferAssetsHandlerV1, startTime) {
		return
	}

	// Determine wallet provider
	walletProvider, err := c.getWalletProvider(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest,
			constants.ErrInvalidProvider, constants.ErrInvalidProvider, err, constants.InitTransferAssetsHandlerV1, startTime)
		return
	}

	// Generate signing payloads (does not execute actual transfer)
	signingPayload, err := walletProvider.InitTransferAssets(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrInitTransferAssets, constants.ErrInitTransferAssets, err, constants.InitTransferAssetsHandlerV1, startTime)
		return
	}

	// Log successful signing payload creation
	log.Printf("[InitTransferAssetsV1] Signing payloads created successfully for UserID: %s, Challenges: %+v\n",
		requestBody.UserID, signingPayload.Challenge)
	instrumentation.SuccessRequestCounter.WithLabelValues(constants.InitTransferAssetsHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.InitTransferAssetsHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, signingPayload)
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
	transferResponse, err := walletProvider.TransferAssets(requestBody)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError,
			constants.ErrTransferAssets, constants.ErrTransferAssets, err, constants.TransferAssetsHandlerV1, startTime)
		return
	}

	// Log successful asset transfer
	log.Printf("[TransferAssetsV1] Assets successfully transferred! UTR: %s, ReceiverID: %s, Amount: %s %s, Fee: %s, TxHash: %s, Network: %s, ReceiverAddress: %s\n",
		transferResponse.UTR, transferResponse.ReceiverID, transferResponse.Amount, transferResponse.Denom, transferResponse.Fee,
		transferResponse.TxInfo.TxHash, transferResponse.TxInfo.Network, transferResponse.TxInfo.ReceiverAddress)

	instrumentation.SuccessRequestCounter.WithLabelValues(constants.TransferAssetsHandlerV1).Inc()
	instrumentation.SuccessLatency.WithLabelValues(constants.TransferAssetsHandlerV1).Observe(time.Since(startTime).Seconds())

	ctx.JSON(http.StatusOK, transferResponse)
}
