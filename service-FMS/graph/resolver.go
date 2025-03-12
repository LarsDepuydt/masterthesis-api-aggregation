package graph

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/LarsDepuydt/masterthesis-api-aggregation/service-FMS/graph/model"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	BuildingsData map[string]*model.Building
}

func LoadBuildingData() map[string]*model.Building {
	// Open the CSV file
	f, err := os.Open("./TMV25.csv")
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV file: %v", err)
	}
	if len(records) < 1 {
		log.Println("CSV file is empty")
		return nil
	}

	// Get header indices
	header := records[0]
	idxGruppe := indexOf(header, "Gruppe")
	idxBygning := indexOf(header, "Bygning")
	idxEtage := indexOf(header, "Etage")
	idxRumnummer := indexOf(header, "Rumnummer")
	idxAreal := indexOf(header, "Areal")
	idxOmkreds := indexOf(header, "Omkreds")
	idxEjendom := indexOf(header, "Ejendom")
	idxBynavn := indexOf(header, "Bynavn")

	buildingsData := make(map[string]*model.Building)

	// Process each CSV row (skip header)
	for i, row := range records[1:] {
		buildingAddress := row[idxBygning]
		floorName := row[idxEtage]
		roomNumber := row[idxRumnummer]

		// Convert area and circumference values
		area, err := strconv.ParseFloat(row[idxAreal], 64)
		if err != nil {
			log.Printf("Error parsing area on row %d: %v", i+2, err)
			continue
		}
		circumference, err := strconv.ParseFloat(row[idxOmkreds], 64)
		if err != nil {
			log.Printf("Error parsing circumference on row %d: %v", i+2, err)
			continue
		}

		// Create or retrieve the Building
		b, ok := buildingsData[buildingAddress]
		if !ok {
			b = &model.Building{
				ID:       buildingAddress, // using address as the unique ID
				Address:  buildingAddress,
				City:     row[idxBynavn],
				Property: row[idxEjendom],
				Floors:   []*model.Floor{},
			}
			buildingsData[buildingAddress] = b
		}

		// Create or retrieve the Floor within the Building
		var floor *model.Floor
		for _, f := range b.Floors {
			if f.Name == floorName {
				floor = f
				break
			}
		}
		if floor == nil {
			floor = &model.Floor{
				ID:           fmt.Sprintf("%s-%s", buildingAddress, floorName),
				Name:         floorName,
				FloorplanURL: "", // Placeholder URL; modify if needed
				Rooms:        []*model.Room{},
			}
			b.Floors = append(b.Floors, floor)
		}

		// Create the Room
		room := &model.Room{
			ID:            fmt.Sprintf("%s-%s", floor.ID, roomNumber),
			Name:          roomNumber,
			Type:          row[idxGruppe],
			Area:          area,
			Circumference: circumference,
		}
		floor.Rooms = append(floor.Rooms, room)
	}

	return buildingsData
}

// indexOf returns the index of target in slice or -1 if not found.
func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}
