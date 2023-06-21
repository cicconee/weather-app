package state

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) SelectEntity(ctx context.Context, stateID string) (Entity, error) {
	e := Entity{ID: stateID}
	return e, e.Select(ctx, s.DB)
}

func (s *Store) InsertEntity(ctx context.Context, state Entity) (sql.Result, error) {
	return state.Insert(ctx, s.DB)
}

func (s *Store) InsertZone(ctx context.Context, zone Zone) error {
	zoneEntity := zone.ToEntity()
	if err := zoneEntity.Insert(ctx, s.DB); err != nil {
		return fmt.Errorf("failed to insert ZoneEntity into database: %w", err)
	}

	for _, polygon := range zone.Geometry {
		perimeterEntity := PerimeterEntity{
			ZoneID: zoneEntity.ID,
			Points: polygon.Permiter(),
		}
		if err := perimeterEntity.Insert(ctx, s.DB); err != nil {
			return fmt.Errorf("failed to insert PerimeterEntity into database: %w", err)
		}

		for _, hole := range polygon.Holes() {
			holeEntity := HoleEntity{
				PerimieterID: perimeterEntity.ID,
				Points:       hole,
			}
			if _, err := holeEntity.Insert(ctx, s.DB); err != nil {
				return fmt.Errorf("failed to insert HoleEntity into database: %w", err)
			}
		}
	}

	return nil
}
