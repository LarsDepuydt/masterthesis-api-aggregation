package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"fmt"
	"io"
	"encoding/json"
	"strings"
	"time"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/joho/godotenv"
)

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

// GraphQL Object Resolvers
type Value struct {
	Timestamp graphql.Time
	Value     float64
}

type Sensor struct {
	ExternalID string
	SourcePath string
	Unit       string
	Type       string
}

type Room struct {
	ID string
}

type Floor struct {
	ID string
}

type Building struct {
	ID string
}

// Query Resolver
type Resolver struct{}

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

// SplitSourcePath extracts structured information from the given source path.
func SplitSourcePath(sourcePath string) (string, string, string, string, error) {
	// Remove leading slash
	sourcePath = strings.TrimPrefix(sourcePath, "/")

	// Split the first part of the path
	parts := strings.SplitN(sourcePath, "/", 2)
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid source path: %s", sourcePath)
	}

	roomSpecification := parts[0]
	restPath := parts[1]

	// Split room specification into max 3 parts
	var buildingID, floorID, roomID string
	roomSpecParts := strings.SplitN(roomSpecification, "_", 3)

	switch len(roomSpecParts) {
	case 1: // Only one part found → It's the building ID
		buildingID = roomSpecParts[0]
		floorID = "undefined"
		roomID = "undefined"
	case 2: // Two parts found → Floor is "undefined", take full second part as room ID
		buildingID = roomSpecParts[0]
		floorID = "undefined"
		roomID = roomSpecParts[1]
	case 3: // Normal case → Assign all three values correctly
		buildingID = roomSpecParts[0]
		floorID = roomSpecParts[1]
		roomID = roomSpecParts[2]
	default:
		return "", "", "", "", fmt.Errorf("unexpected error while parsing room specification: %s", roomSpecification)
	}

	// Extract sensorType (last element of restPath)
	restPathParts := strings.Split(restPath, "/")
	sensorType := restPathParts[len(restPathParts)-1]

	return buildingID, floorID, roomID, sensorType, nil
}

// Resolver for Building.Floors
func (b *Building) Floors(ctx context.Context, args struct{ FloorIDs *[]string }) ([]*Floor, error) {
	metadata, err := FetchMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	// Collect unique floors for this building
	floorMap := make(map[string]*Floor)
	for _, meta := range metadata {
		buildingID, floorID, _, _, err := SplitSourcePath(meta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source path: %v", err)
		}

		if buildingID != b.ID {
			continue
		}

		if _, exists := floorMap[floorID]; !exists {
			floorMap[floorID] = &Floor{ID: floorID}
		}
	}

	// Apply floor filter
	var floors []*Floor
	for _, floor := range floorMap {
		if args.FloorIDs == nil || len(*args.FloorIDs) == 0 {
			floors = append(floors, floor)
		} else {
			for _, id := range *args.FloorIDs {
				if floor.ID == id {
					floors = append(floors, floor)
					break
				}
			}
		}
	}

	return floors, nil
}

// Resolver for Floor.Rooms
func (f *Floor) Rooms(ctx context.Context, args struct{ RoomIDs *[]string }) ([]*Room, error) {
	metadata, err := FetchMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	// Collect unique rooms for this floor
	roomMap := make(map[string]*Room)
	for _, meta := range metadata {
		buildingID, floorID, roomID, _, err := SplitSourcePath(meta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source path: %v", err)
		}

		if buildingID != "" && floorID != f.ID {
			continue
		}

		if _, exists := roomMap[roomID]; !exists {
			roomMap[roomID] = &Room{ID: roomID}
		}
	}

	// Apply room filter
	var rooms []*Room
	for _, room := range roomMap {
		if args.RoomIDs == nil || len(*args.RoomIDs) == 0 {
			rooms = append(rooms, room)
		} else {
			for _, id := range *args.RoomIDs {
				if room.ID == id {
					rooms = append(rooms, room)
					break
				}
			}
		}
	}

	return rooms, nil
}

// Resolver for Room.Sensors
func (r *Room) Sensors(ctx context.Context, args struct{ SensorIDs *[]string }) ([]*Sensor, error) {
	metadata, err := FetchMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	// Collect sensors for this room
	var sensors []*Sensor
	for _, meta := range metadata {
		_, _, roomID, sensorType, err := SplitSourcePath(meta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source path: %v", err)
		}

		// Skip sensors that don't belong to this room
		if roomID != r.ID {
			continue
		}

		// Convert ExternalID (int32) to a string
		externalIDStr := fmt.Sprintf("%d", meta.ExternalID)

		// Create the Sensor object
		sensor := &Sensor{
			ExternalID: externalIDStr, // Use the string version of ExternalID
			SourcePath: meta.Source,
			Unit:       meta.Unit,
			Type:       sensorType,
		}

		// Apply sensor filter
		if args.SensorIDs == nil || len(*args.SensorIDs) == 0 {
			// If no filter is provided, include all sensors
			sensors = append(sensors, sensor)
		} else {
			// Include only sensors with matching IDs
			for _, id := range *args.SensorIDs {
				if sensor.ExternalID == id {
					sensors = append(sensors, sensor)
					break
				}
			}
		}
	}

	return sensors, nil
}

// Resolver for Sensor.Values
func (s *Sensor) Values(ctx context.Context, args struct {
	StartTime graphql.Time
	EndTime   *graphql.Time
}) ([]Value, error) {
	log.Println("Fetching values for sensor:", s.ExternalID)
	log.Println("Start Time:", args.StartTime)

	startTime := args.StartTime.Time

	var endTime time.Time
	if args.EndTime != nil {
		endTime = args.EndTime.Time
	} else {
		endTime = time.Now().UTC()
	}

	url := fmt.Sprintf("https://bms-api.build.aau.dk/api/v1/trenddata?externallogid=%s&starttime=%s&endtime=%s",
		s.ExternalID,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)
	body, err := SendRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trend data: %v", err)
	}

	var trendData []TrendDataResponse
	if err := json.Unmarshal(body, &trendData); err != nil {
		return nil, fmt.Errorf("failed to parse trend data JSON: %v", err)
	}

	// Convert timestamps and map to Value struct
	var values []Value
	for _, data := range trendData {
		t, err := time.Parse("2006-01-02 15:04:05", data.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert timestamp: %v", err)
		}

		values = append(values, Value{
			Timestamp: graphql.Time{Time: t},
			Value:     data.Value,
		})
	}

	return values, nil
}

// Resolver for Query.Building
func (r *Resolver) Building(ctx context.Context, args struct{ IDs *[]string }) ([]*Building, error) {
	metadata, err := FetchMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	// Collect unique buildings
	buildingMap := make(map[string]*Building)
	for _, meta := range metadata {
		buildingID, _, _, _, err := SplitSourcePath(meta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse source path: %v", err)
		}

		if _, exists := buildingMap[buildingID]; !exists {
			buildingMap[buildingID] = &Building{ID: buildingID}
		}
	}

	// Apply building filter
	var buildings []*Building
	for _, building := range buildingMap {
		if args.IDs == nil || len(*args.IDs) == 0 {
			buildings = append(buildings, building)
		} else {
			for _, id := range *args.IDs {
				if building.ID == id {
					buildings = append(buildings, building)
					break
				}
			}
		}
	}

	return buildings, nil
}

func main() {
	godotenv.Load()

	// Read GraphQL schema
	schemaFile, err := os.ReadFile("BMS-api.schema.graphql")
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
