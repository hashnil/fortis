package main

import (
	"fortis/app"
	"fortis/infrastructure/config"
	"log"
	"os"
)

func main() {
	// Bootstrap configuration
	if err := config.LoadConfig(); err != nil {
		log.Println("failed to load configurations: ", err)
		os.Exit(1)
	}

	// Start the application
	svc, err := app.NewService()
	if err != nil {
		log.Println("failed to start wallet-service [FORTIS]: ", err)
		os.Exit(1)
	}

	// Start the service
	svc.Run()
}
