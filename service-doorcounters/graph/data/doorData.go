package data

import "github.com/LarsDepuydt/masterthesis-api-aggregation/service-doorcounters/graph/model"

var (
	// staticEntrances is the single source of truth for entrance definitions
	StaticEntrances = []*model.Entrance{
		{ID: "a", Name: "Door A - direction parking lot"},
		{ID: "b", Name: "Door B - direction Build building"},
		{ID: "c", Name: "Door C - direction campus"},
	}

	// entranceMap provides O(1) lookup by ID
	EntranceMap = func() map[string]*model.Entrance {
		m := make(map[string]*model.Entrance)
		for _, e := range StaticEntrances {
			m[e.ID] = e
		}
		return m
	}()
)
