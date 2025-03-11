package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{}

// Api response
type MetaDataResponse struct {
	ExternalID int32  `json:"externallogid"`
	Source     string `json:"source"`
	Unit       string `json:"unit"`
}

type TrendDataResponse struct {
	ExternalID int32   `json:"externallogid"`
	Timestamp  string  `json:"timestamp"`
	Value      float64 `json:"value"`
}

// Global variable to store metadata
var metadata []MetaDataResponse
var metadataOnce sync.Once

func SendRequest(endpoint string) ([]byte, error) {
	username := os.Getenv("BMS_USERNAME")
	password := os.Getenv("BMS_PASSWORD")

	// Create a new HTTP request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Basic Authentication
	req.SetBasicAuth(username, password)

	// Use http.Client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body after non-200 status: %v", err)
		}
		bodyString := string(bodyBytes) // Convert byte slice to string
		log.Printf("API returned non-200 status: %d, response: %s", resp.StatusCode, bodyString)
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

// FetchMetadata makes an API call to get metadata
func FetchMetaData() ([]MetaDataResponse, error) {
	var err error
	metadataOnce.Do(func() {
		var body []byte
		body, err = SendRequest("https://bms-api.build.aau.dk/api/v1/metadata")
		if err != nil {
			return
		}
		if err = json.Unmarshal(body, &metadata); err != nil {
			return
		}
	})
	return metadata, err
}
