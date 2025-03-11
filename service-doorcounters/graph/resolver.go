package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	baseURL      = "http://iot.multiteknik.dk:8080"
	authEndpoint = "/api/auth/login"
	dataEndpoint = "/api/plugins/telemetry/DEVICE/47afeb80-276e-11ec-92de-537d4a380471/values/timeseries"
)

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct{}

// AuthResponse stores the token received after login
type AuthResponse struct {
	Token string `json:"token"`
}

// TelemetryResponse represents the API response
type TelemetryResponse map[string][]struct {
	Ts    int64  `json:"ts"`
	Value string `json:"value"`
}

// TelemetryData represents the structured telemetry data
type TelemetryData struct {
	Timestamp int64  `json:"timestamp"`
	DoorA     *int32 `json:"c1"`
	DoorB     *int32 `json:"c3"`
	DoorC     *int32 `json:"c2"`
}

// FetchAuthToken fetches the authentication token
func FetchAuthToken() (string, error) {
	username := os.Getenv("DOOR_USERNAME")
	password := os.Getenv("DOOR_PASSWORD")

	authPayload := map[string]string{"username": username, "password": password}
	payloadBytes, _ := json.Marshal(authPayload)

	req, err := http.NewRequest("POST", baseURL+authEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var authResp AuthResponse
	json.Unmarshal(body, &authResp)

	if authResp.Token == "" {
		return "", fmt.Errorf("failed to retrieve auth token")
	}
	return authResp.Token, nil
}

// FetchTelemetryData fetches IoT telemetry data
func FetchTelemetryData(startTs, endTs int64) ([]TelemetryData, error) {
	token, err := FetchAuthToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s?keys=c1,c2,c3&startTs=%d&endTs=%d&limit=1000", baseURL, dataEndpoint, startTs, endTs)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var telemetryResponse TelemetryResponse
	json.Unmarshal(body, &telemetryResponse)

	var data []TelemetryData
	for key, values := range telemetryResponse {
		for _, entry := range values {
			// Convert string to integer
			var valueInt int32
			fmt.Sscanf(entry.Value, "%d", &valueInt)

			// Find existing entry or create a new one
			found := false
			for i := range data {
				if data[i].Timestamp == entry.Ts {
					switch key {
					case "c1": // Indgang A (to parking, street)
						data[i].DoorA = &valueInt
					case "c2": // Indgang C (to campus, bus station)
						data[i].DoorC = &valueInt
					case "c3": // Indgang B (to building)
						data[i].DoorB = &valueInt
					}
					found = true
					break
				}
			}

			// If not found, create a new entry
			if !found {
				newEntry := TelemetryData{
					Timestamp: entry.Ts,
				}
				switch key {
				case "c1":
					newEntry.DoorA = &valueInt
				case "c2":
					newEntry.DoorC = &valueInt
				case "c3":
					newEntry.DoorB = &valueInt
				}
				data = append(data, newEntry)
			}
		}
	}
	return data, nil
}
