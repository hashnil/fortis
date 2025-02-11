package factory

import (
	"fmt"
	"fortis/entity/constants"
	"fortis/internal/integration/db"
	"fortis/internal/integration/db/postgresql"
)

// Database factory
func InitDBClient(dbType string) (db.Client, error) {
	switch dbType {
	case constants.PostgreSQL:
		// Initialize PostgreSQL client for database operations.
		dbClient, err := postgresql.NewPostgresClient()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize postgres client: %w", err)
		}
		return dbClient, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
