package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.70

import (
	"context"
	"fmt"

	"github.com/LarsDepuydt/masterthesis-api-aggregation/service-coffee/graph/model"
)

// FindFloorByID is the resolver for the findFloorByID field.
func (r *entityResolver) FindFloorByID(ctx context.Context, id string) (*model.Floor, error) {
	return &model.Floor{ID: id}, nil
}

// FindMachineByID is the resolver for the findMachineByID field.
func (r *entityResolver) FindMachineByID(ctx context.Context, id string) (*model.Machine, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT machine_id, machine_name, floor_id
		FROM machines
		WHERE machine_id = $1
	`, id)

	var machine model.Machine
	var floorID string
	if err := row.Scan(&machine.ID, &machine.Name, &floorID); err != nil {
		return nil, fmt.Errorf("could not find machine with ID %s: %w", id, err)
	}

	return &machine, nil
}

// Entity returns EntityResolver implementation.
func (r *Resolver) Entity() EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
