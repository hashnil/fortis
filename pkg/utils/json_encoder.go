package utils

import (
	"encoding/json"
	"log"
)

// MarshalToJSON converts a struct to a JSON byte array.
func MarshalToJSON(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to marshal data: %v", err)
	}
	return jsonData
}
