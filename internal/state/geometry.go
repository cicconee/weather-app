package state

import (
	"context"

	"github.com/cicconee/weather-app/internal/geometry"
)

type Geometry []Perimeter

func NewGeometry(mp geometry.MultiPolygon) Geometry {
	g := Geometry{}

	for _, polygon := range mp {
		g = append(g, NewPerimeter(polygon))
	}

	return g
}

type Perimeter struct {
	ID     int
	ZoneID int
	Points geometry.PointCollection
	Holes  HoleCollection
}

func NewPerimeter(poly geometry.Polygon) Perimeter {
	p := Perimeter{
		Points: poly.Permiter(),
		Holes:  NewHoleCollection(poly.Holes()),
	}

	return p
}

func (p *Perimeter) Insert(ctx context.Context, db QueryRower) error {
	query := `
		INSERT INTO state_zone_perimeters(sz_id, boundary)
		VALUES($1, $2)
		RETURNING id`

	if err := db.QueryRowContext(ctx, query,
		p.ZoneID,
		p.Points.String(),
	).Scan(&p.ID); err != nil {
		return err
	}

	for _, hole := range p.Holes {
		hole.PerimieterID = p.ID

		if err := hole.Insert(ctx, db); err != nil {
			return nil
		}
	}

	return nil
}

type HoleCollection []Hole

func NewHoleCollection(geoHoles []geometry.PointCollection) HoleCollection {
	h := HoleCollection{}

	for i := range geoHoles {
		h = append(h, Hole{Points: geoHoles[i]})
	}

	return h
}

type Hole struct {
	ID           int
	PerimieterID int
	Points       geometry.PointCollection
}

func (h *Hole) Insert(ctx context.Context, db QueryRower) error {
	query := `
		INSERT INTO state_zone_holes(zp_id, boundary)
		VALUES($1, $2)
		RETURNING id`

	return db.QueryRowContext(ctx, query,
		h.PerimieterID,
		h.Points.String(),
	).Scan(&h.ID)
}
