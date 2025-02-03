package app

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"fortis/api/controller"
	"fortis/internal/instrumentation"
)

type Service struct {
	Engine *gin.Engine // Gin Engine for handling HTTP requests
}

var (
	engineLock sync.RWMutex // Mutex for ensuring thread-safe engine initialization
	engine     *gin.Engine  // Singleton instance of the Gin engine
)

// NewService initializes the Service struct, sets up the Gin engine, and registers routes.
func NewService() (*Service, error) {
	healthController := controller.NewHealthController()
	walletController := controller.NewWalletController()
	// if err != nil {
	// 	return &Service{}, fmt.Errorf("unable to create controller: %v", err)
	// }

	// Ensure the engine is initialized only once
	if engine == nil {
		initializeEngine()
	}

	// Register application routes
	registerRoutes(healthController, walletController)

	// Start prometheus server
	instrumentation.StartPrometheusServer()

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
