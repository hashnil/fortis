package postgresql

import (
	"fmt"
	"fortis/internal/integration/db"
	dbmodel "fortis/internal/integration/db/models"
	"log"

	"github.com/getpanda/commons/db/postgresql/dsn"
	"gorm.io/gorm"
)

type PostgresSQLClient struct {
	client *gorm.DB
}

func NewPostgresClient() (db.Client, error) {
	// Initialize the PostgresSQLClient from commons
	db, err := dsn.ConnectWithConnectorGorm()
	if err != nil {
		return &PostgresSQLClient{}, err
	}

	// Create the tables if it does not exist
	if err := db.AutoMigrate(&dbmodel.Wallet{}, &dbmodel.TransactionLog{}); err != nil {
		return &PostgresSQLClient{}, fmt.Errorf("error creating tables: %v", err)
	}

	log.Println("Postgres SQL client initialized successfully.")
	return &PostgresSQLClient{client: db}, nil
}
