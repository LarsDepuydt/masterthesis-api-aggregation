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

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/joho/godotenv"
)

// Api response
type MetaDataResponse struct {
	ExternalID int32 `json:"externallogid"`
	Source string `json:"source"`
	Unit string `json:"unit"`
}

type TrendDataResponse struct {
	ExternalID int32 `json:"externallogid"`
	Timestamp string `json:"timestamp"`
	Value float64 `json:"value"`
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
	ID      string
	Sensors []Sensor
}

type Floor struct {
	ID    string
	Rooms []Room
}

type Building struct {
	ID     string
	Floors []Floor
}

// Query Resolver
type Resolver struct{}

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
	    bodyString := string(bodyBytes)  // Convert byte slice to string
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
	body, err := SendRequest("https://bms-api.build.aau.dk/api/v1/metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to send metadata request: %v", err)
	}

	var metadata []MetaDataResponse
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %v", err)
	}

	return metadata, nil
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

func ParseMetaData(metaData []MetaDataResponse, filterIDs *[]string) ([]Building, error) {
	buildingMap := make(map[string]*Building)

	for _, meta := range metaData {
		buildingID, floorID, roomID, sensorType, err := SplitSourcePath(meta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse room specification: %v", err)
		}

		// Find or create the building
		building, exists := buildingMap[buildingID]
		if !exists {
			building = &Building{ID: buildingID, Floors: []Floor{}}
			buildingMap[buildingID] = building
		}

		// Find or create the floor
		var floor *Floor
		for i := range building.Floors {
			if building.Floors[i].ID == floorID {
				floor = &building.Floors[i]
				break
			}
		}
		if floor == nil {
			building.Floors = append(building.Floors, Floor{ID: floorID, Rooms: []Room{}})
			floor = &building.Floors[len(building.Floors)-1]
		}

		// Find or create the room
		var room *Room
		for i := range floor.Rooms {
			if floor.Rooms[i].ID == roomID {
				room = &floor.Rooms[i]
				break
			}
		}
		if room == nil {
			floor.Rooms = append(floor.Rooms, Room{ID: roomID, Sensors: []Sensor{}})
			room = &floor.Rooms[len(floor.Rooms)-1]
		}

		// Add sensor to the room
		sensor := Sensor{
			ExternalID: fmt.Sprintf("%d", meta.ExternalID),
			SourcePath: meta.Source,
			Unit:       meta.Unit,
			Type:       sensorType,
		}
		room.Sensors = append(room.Sensors, sensor)
	}

	// Convert map values to slice and filter by IDs
	var buildings []Building
	if filterIDs == nil || len(*filterIDs) == 0 {
		// If no filter is provided, return all buildings
		for _, b := range buildingMap {
			buildings = append(buildings, *b)
		}
	} else {
		// Return only buildings that match the filter
		idSet := make(map[string]struct{}, len(*filterIDs))
		for _, id := range *filterIDs {
			idSet[id] = struct{}{}
		}
		for _, b := range buildingMap {
			if _, exists := idSet[b.ID]; exists {
				buildings = append(buildings, *b)
			}
		}
	}

	return buildings, nil
}

func (r *Resolver) Building(ctx context.Context, args struct{ IDs *[]string }) ([]Building, error) {
	log.Println("Fetching building with IDs:", args.IDs)

	metadata, err := FetchMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	buildings, err := ParseMetaData(metadata, args.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	return buildings, nil
}

func convertToTime(input string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", input)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %v", err)
	}
	return t, nil
}

func (s Sensor) Values(ctx context.Context, args struct {
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

func main() {
	godotenv.Load()
	// metadataURL := "https://bms-api.build.aau.dk/api/v1/metadata"
	// trendDataURL := "https://bms-api.build.aau.dk/api/v1/trenddata"
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
