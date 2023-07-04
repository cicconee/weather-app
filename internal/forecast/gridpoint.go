package forecast

import (
	"context"
	"database/sql"

	"github.com/cicconee/weather-app/internal/geometry"
)

// GridpointAPIResource is the gridpoint data that is returned by ForecastAPI.
// GridpointAPIResource should never be explicitly created and only be used when
// returned from ForecastAPI.
//
// A GridpointAPIResource represents a 2.5km grid square that can be used to get
// forecast data.
//
// GridpointAPIResource can be converted into a GridpointEntity by calling
// ToGridpointEntity.
type GridpointAPIResource struct {
	// The three-letter identifier for a NWS office. This identifies the grid.
	GridID string `json:"gridId"`

	// The x coordinate in the grid.
	GridX int `json:"gridX"`

	// The y coordinate in the grid.
	GridY int `json:"gridY"`

	// The timezone used in the grid.
	TimeZone string `json:"timeZone"`
}

// ToGridpointEntity returns this GridpointAPIResource as a GridpointEntity.
// Only the GridID, GridX, GridY, and TimeZone fields are populated in the
// returned GridpointEntity.
//
// The GridpointEntity will need to have its Timeline and Geometry set.
func (g *GridpointAPIResource) ToGridpointEntity() GridpointEntity {
	return GridpointEntity{
		GridID:   g.GridID,
		GridX:    g.GridX,
		GridY:    g.GridY,
		TimeZone: g.TimeZone,
	}
}

// GridpointEntity is a gridpoint database entity. Each gridpoint will have a
// unique GridID, GridX, GridY combination. GridpointEntity is identified by
// ID in the database.
//
// GridpointEntity should only be written to the database if it was returned
// by the ToGridpointEntity of a GridpointAPIResource. Do not set the GridID,
// GridX, GridY, and TimeZone fields once returned.
//
// All period database entities depend on a gridpoint. A period cannot exist
// without a gridpoint.
type GridpointEntity struct {
	// The database identifier. ID is set after writing or reading a
	// GridpointEntity to the database.
	ID int

	// The grid identifier.
	GridID string

	// The grids x coordinate.
	GridX int

	// The grids y coordinate.
	GridY int

	// The time zone used in the gridpoint.
	TimeZone string

	// The time of generation and expiration of the gridpoints forecast data.
	Timeline Timeline

	// The geographical boundary that this gridpoint covers. Any coordinate
	// that resides within this polygon will get its forecast data from this
	// gridpoint.
	Geometry geometry.Polygon
}

// Scan will scan the query result in scanner into this GridpointEntity.
func (g *GridpointEntity) Scan(scanner Scanner) error {
	return scanner.Scan(
		&g.ID,
		&g.GridID,
		&g.GridX,
		&g.GridY,
		&g.Timeline.GeneratedAt,
		&g.Timeline.ExpiresAt,
		&g.TimeZone)
}

// Select reads a gridpoint into this GridpointEntity where point resides inside
// its geometric bounds.
func (g *GridpointEntity) Select(ctx context.Context, db *sql.DB, point geometry.Point) error {
	query := `SELECT id, grid_id, grid_x, grid_y, generated_at, expires_at, timezone
			  FROM gridpoints WHERE boundary @> $1`

	return g.Scan(db.QueryRowContext(ctx, query, point.RoundedString()))
}

// Insert writes this GridpointEntity into the database and sets this
// GridpointEntity ID field.
func (g *GridpointEntity) Insert(ctx context.Context, db QueryRower) error {
	query := `INSERT INTO gridpoints(grid_id, grid_x, grid_y, generated_at, expires_at, timezone, 
			  boundary) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	return db.QueryRowContext(ctx, query,
		g.GridID,
		g.GridX,
		g.GridY,
		g.Timeline.GeneratedAt,
		g.Timeline.ExpiresAt,
		g.TimeZone,
		g.Geometry.Permiter().String()).Scan(&g.ID)
}

// Update writes this GridpointEntity to the database as an update. Only the Timeline
// can be updated.
//
// The only fields that need to be set are the ID and Timeline.
func (g *GridpointEntity) Update(ctx context.Context, db Execer) error {
	query := `UPDATE gridpoints SET generated_at = $1, expires_at = $2
			  WHERE id = $3`

	_, err := db.ExecContext(ctx, query,
		g.Timeline.GeneratedAt,
		g.Timeline.ExpiresAt,
		g.ID)

	return err
}
