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

type Resolver struct {
	metadata     []MetaDataResponse
	metadataOnce sync.Once
}

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

// FetchMetadata makes an API call to get metadata.
// It checks if metadata is already populated on the Resolver struct.
// The sync.Once ensures that the API call and unmarshaling happen only once
// even if multiple goroutines call this concurrently when metadata is empty.
func (r *Resolver) FetchMetaData() ([]MetaDataResponse, error) {
	// First, check if metadata is already loaded.
	// This non-blocking check avoids acquiring the sync.Once mutex
	// if the data is already there.
	if r.metadata != nil && len(r.metadata) > 0 {
		return r.metadata, nil
	}

	var fetchErr error // Declare fetchErr to capture error from the Do function

	// Use sync.Once to ensure the fetching logic runs only once,
	// especially for the first time or if a previous attempt failed (and cleared metadata).
	r.metadataOnce.Do(func() {
		log.Println("FetchMetaData: Metadata not loaded. Initiating API call...")
		var body []byte
		body, fetchErr = SendRequest("https://bms-api.build.aau.dk/api/v1/metadata")
		if fetchErr != nil {
			log.Printf("FetchMetaData: Error sending request: %v", fetchErr)
			// Do not return here, let fetchErr be set and returned by the outer function.
			return
		}

		// Unmarshal into the Resolver's metadata field
		if fetchErr = json.Unmarshal(body, &r.metadata); fetchErr != nil {
			log.Printf("FetchMetaData: Error unmarshaling response: %v", fetchErr)
			// Do not return here, let fetchErr be set and returned by the outer function.
			return
		}
		log.Printf("FetchMetaData: Successfully loaded %d metadata entries.", len(r.metadata))
	})

	// If fetchErr was set inside the Do function, return it.
	// Otherwise, return the loaded metadata.
	return r.metadata, fetchErr
}
