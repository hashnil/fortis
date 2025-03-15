package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/getpanda/commons/db/redis"
	"github.com/getpanda/commons/pkg/auth/core"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"fortis/api/controller"
	"fortis/entity/constants"
	"fortis/infrastructure/config"
	"fortis/infrastructure/factory"
	"fortis/internal/instrumentation"
)

type Service struct {
	Engine *gin.Engine // Gin Engine for handling HTTP requests
}

var (
	engineLock sync.RWMutex // Mutex for ensuring thread-safe engine initialization
	engine     *gin.Engine  // Singleton instance of the Gin engine

	authMiddleware *core.AuthMiddleware // Used for JWT & Nonce authentication
)

// NewService initializes the Service struct, sets up the Gin engine, and registers routes.
func NewService() (*Service, error) {
	// Initialize PostgreSQL client for database operations.
	dbClient, err := factory.InitDBClient(constants.PostgreSQL)
	if err != nil {
		return &Service{}, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	// Initialize redis client
	redisClient, err := redis.InitRedisClient(context.Background(), viper.GetInt("db.redis.index"), viper.GetInt("db.redis.pool_size"))
	if err != nil {
		return &Service{}, fmt.Errorf("failed to initialize redis client: %v", err)
	}

	// Start prometheus server
	instrumentation.StartPrometheusServer()

	// Initialize JWT and Nonce validators
	jwtValidator := core.NewJWTValidator(string(config.JWT_SECRET))
	nonceValidator, err := core.NewNonceValidator(redisClient, string(config.NONCE_SECRET))
	if err != nil {
		return &Service{}, fmt.Errorf("unable to create nonce validator: %v", err)
	}
	authMiddleware = core.NewAuthMiddleware(jwtValidator, nonceValidator)

	// Initialize controllers
	healthController := controller.NewHealthController()
	walletController, err := controller.NewWalletController(dbClient)
	if err != nil {
		return &Service{}, fmt.Errorf("unable to create controller: %v", err)
	}

	// Ensure the engine is initialized only once
	if engine == nil {
		initializeEngine()
	}

	// Register application routes
	registerRoutes(healthController, walletController)

	return &Service{
		Engine: engine,
	}, nil
}

// initializeEngine initializes the Gin engine in a thread-safe manner.
func initializeEngine() {
	engineLock.Lock()
	defer engineLock.Unlock()

	if engine == nil {
		engine = gin.New()
		engine.Use(gin.Recovery(), gin.Logger())
		gin.SetMode(gin.ReleaseMode)         // Use ReleaseMode for production
		engine.HandleMethodNotAllowed = true // Enable 405 Method Not Allowed responses
	}
}

// Run starts the HTTP server on port 8080.
func (s *Service) Run() error {
	return s.Engine.Run(":" + viper.GetString("service.port"))
}
