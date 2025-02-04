package postgresql

import (
	"fmt"
	"fortis/internal/integration/db"
	dbmodel "fortis/internal/integration/db/models"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresSQLClient struct {
	client *gorm.DB
}

func NewPostgresClient() (db.Client, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		viper.GetString("db.postgres.host"),
		viper.GetString("db.postgres.user"),
		viper.GetString("db.postgres.pass"),
		viper.GetString("db.postgres.name"),
		viper.GetString("db.postgres.port"),
	)
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return &PostgresSQLClient{}, err
	}

	// Create the wallet table if it does not exist
	if err := db.AutoMigrate(&dbmodel.Wallet{}); err != nil {
		return &PostgresSQLClient{}, fmt.Errorf("error creating tables: %v", err)
	}

	log.Println("Postgres SQL client initialized successfully.")
	return &PostgresSQLClient{client: db}, nil
}
