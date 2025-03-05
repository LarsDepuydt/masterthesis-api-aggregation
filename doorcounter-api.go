package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/joho/godotenv"
)

const (
	baseURL      = "http://iot.multiteknik.dk:8080"
	authEndpoint = "/api/auth/login"
	dataEndpoint = "/api/plugins/telemetry/DEVICE/47afeb80-276e-11ec-92de-537d4a380471/values/timeseries"
)

// AuthResponse stores the token received after login
type AuthResponse struct {
	Token string `json:"token"`
}

// TelemetryResponse represents the API response
type TelemetryResponse map[string][]struct {
	Ts    int64  `json:"ts"`
	Value string `json:"value"`
}

// GetAuthToken fetches the authentication token
func GetAuthToken() (string, error) {
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

	body, _ := ioutil.ReadAll(resp.Body)
	var authResp AuthResponse
	json.Unmarshal(body, &authResp)

	if authResp.Token == "" {
		return "", fmt.Errorf("failed to retrieve auth token")
	}
	return authResp.Token, nil
}

// GetTelemetryData fetches IoT telemetry data
func GetTelemetryData(startTs, endTs int64) ([]TelemetryData, error) {
	token, err := GetAuthToken()
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

	body, _ := ioutil.ReadAll(resp.Body)
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

type TelemetryData struct {
	Timestamp int64  `json:"timestamp"`
	DoorA     *int32 `json:"c1"`
	DoorB     *int32 `json:"c3"`
	DoorC     *int32 `json:"c2"`
}

type TelemetryDataGraphQL struct {
	Timestamp string `json:"timestamp"`
	DoorA     *int32 `json:"doorA"`
	DoorB     *int32 `json:"doorB"`
	DoorC     *int32 `json:"doorC"`
}

// Resolver struct
type Resolver struct{}

func (r *Resolver) GetTelemetryData(ctx context.Context, args struct {
	StartTime graphql.Time
	EndTs     *graphql.Time // Optional end time
}) ([]TelemetryDataGraphQL, error) {
	// Convert GraphQL Time to int64 (UNIX timestamp in seconds)
	startTs := args.StartTime.Time.Unix() * 1000

	var endTs int64
	if args.EndTs != nil {
		endTs = args.EndTs.Time.Unix() * 1000
	} else {
		endTs = time.Now().Unix() * 1000 // Default to current time if not provided
	}

	data, err := GetTelemetryData(startTs, endTs)
	if err != nil {
		log.Println("Error fetching telemetry data:", err)
		return nil, err
	}

	// Convert timestamps to ISO 8601 format for GraphQL response
	var graphqlData []TelemetryDataGraphQL
	for _, d := range data {
		graphqlData = append(graphqlData, TelemetryDataGraphQL{
			Timestamp: time.Unix(d.Timestamp/1000, 0).Format(time.RFC3339),
			DoorA:     d.DoorA,
			DoorB:     d.DoorB,
			DoorC:     d.DoorC,
		})
	}

	return graphqlData, nil
}

func main() {
	godotenv.Load()

	// Read GraphQL schema
	schemaFile, err := os.ReadFile("doorcounter-api.schema.graphql")
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	// Parse schema
	schema := graphql.MustParseSchema(
		string(schemaFile),
		&Resolver{},
		graphql.UseFieldResolvers(),
	)

	// Start server
	http.Handle("/graphql", &relay.Handler{Schema: schema})
	log.Println("🚀 GraphQL server running on http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
