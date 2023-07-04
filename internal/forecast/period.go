package forecast

import (
	"context"
	"database/sql"
	"sort"
	"time"
)

// Period is the weather data for a 1-hour period of time. The Number field
// corresponds to the point in time this period belongs. The StartTime and
// EndTime is formatted to the local timezone of the area that the Period is
// reporting for. A Period is safe to be consumed by external packages.
//
// When periods are in a collection, organizing them in ascending order by
// number corresponds to moving forward in time.
type Period struct {
	Number          int       `json:"number"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	IsDaytime       bool      `json:"isDaytime"`
	Temperature     int       `json:"temperature"`
	TemperatureUnit string    `json:"temperatureUnit"`
	WindSpeed       string    `json:"windSpeed"`
	WindDirection   string    `json:"windDirection"`
	ShortForecast   string    `json:"shortForecast"`
}

// loadTimeZone formats the StartTime and EndTime of this Period to loc.
func (p *Period) loadTimeZone(loc *time.Location) {
	p.StartTime = p.StartTime.In(loc)
	p.EndTime = p.EndTime.In(loc)
}

// PeriodCollection is a collection of Period. PeriodCollection will be
// sorted in ascending order by the Number field of a Period. To verify it
// is sorted use the method IsSorted. If for any reason the PeriodCollection
// is not sorted, use the Sort method.
type PeriodCollection []Period

// loadTimeZone formats the StartTime and EndTime of each Period to loc.
func (p *PeriodCollection) loadTimeZone(loc *time.Location) {
	for i := range *p {
		(*p)[i].loadTimeZone(loc)
	}
}

// IsSorted verifys that this PeriodCollection is sorted by the Number field
// of Period.
func (p *PeriodCollection) IsSorted() bool {
	return sort.SliceIsSorted(*p, func(i, j int) bool {
		return (*p)[i].Number < (*p)[j].Number
	})
}

// Sort sorts this PeriodCollection by the Number field of Period. Sort will
// call the IsSorted method before sorting. There is no need to check if it
// is sorted before calling this method.
func (p *PeriodCollection) Sort() {
	if p.IsSorted() {
		return
	}

	sort.Slice(*p, func(i, j int) bool {
		return (*p)[i].Number < (*p)[i].Number
	})
}

// PeriodAPIResource is the 1-hour weather data of a forecast that is returned
// by ForecastAPI. PeriodAPIResource should never be explicitly created and only
// be used when returned from ForecastAPI.
//
// A period represents a 1-hour period in a forecast. Each period holds
// meteorological data for this 1-hour period. Periods are organized in ascending
// order by the Number field to create a valid hourly forecast.
//
// PeriodAPIResource can be converted into a PeriodEntity by calling ToPeriodEntity.
type PeriodAPIResource struct {
	Number          int       `json:"number"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	IsDaytime       bool      `json:"isDaytime"`
	Temperature     int       `json:"temperature"`
	TemperatureUnit string    `json:"temperatureUnit"`
	WindSpeed       string    `json:"windSpeed"`
	WindDirection   string    `json:"windDirection"`
	ShortForecast   string    `json:"shortForecast"`
}

// ToPeriodEntity returns this PeriodAPIResource as a PeriodEntity.
func (p *PeriodAPIResource) ToPeriodEntity() PeriodEntity {
	return PeriodEntity{
		Number:          p.Number,
		StartTime:       p.StartTime.UTC(),
		EndTime:         p.EndTime.UTC(),
		IsDaytime:       p.IsDaytime,
		Temperature:     p.Temperature,
		TemperatureUnit: p.TemperatureUnit,
		WindSpeed:       p.WindSpeed,
		WindDirection:   p.WindDirection,
		ShortForecast:   p.ShortForecast,
	}
}

// PeriodEntity is a period database entity. Each period will have a unique
// Number GridpointID combination and this will be its identifier.
//
// PeriodEntity should only be written to the database if it was returned
// by the ToPeriodEntity method of a PeriodAPIResource.
//
// A period belongs to a gridpoint. It cannot exist without a gridpoint.
type PeriodEntity struct {
	Number          int
	StartTime       time.Time
	EndTime         time.Time
	IsDaytime       bool
	Temperature     int
	TemperatureUnit string
	WindSpeed       string
	WindDirection   string
	ShortForecast   string
	GridpointID     int
}

// ToPeriod returns this PeriodEntity as a Period.
func (p *PeriodEntity) ToPeriod() Period {
	return Period{
		Number:          p.Number,
		StartTime:       p.StartTime,
		EndTime:         p.EndTime,
		IsDaytime:       p.IsDaytime,
		Temperature:     p.Temperature,
		TemperatureUnit: p.TemperatureUnit,
		WindSpeed:       p.WindSpeed,
		WindDirection:   p.WindDirection,
		ShortForecast:   p.ShortForecast,
	}
}

// Scan will scan the query result in scanner into this PeriodEntity.
func (p *PeriodEntity) Scan(scanner Scanner) error {
	return scanner.Scan(
		&p.Number,
		&p.StartTime,
		&p.EndTime,
		&p.IsDaytime,
		&p.Temperature,
		&p.TemperatureUnit,
		&p.WindSpeed,
		&p.WindDirection,
		&p.ShortForecast,
		&p.GridpointID)
}

// Insert writes this PeriodEntity into the database. All fields being written
// must be set before calling this method.
func (p *PeriodEntity) Insert(ctx context.Context, db *sql.DB) error {
	query := `INSERT INTO periods(num, starts, ends, is_day_time, temp, temp_unit, wind_speed,
			  wind_direction, short_forecast, gp_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := db.ExecContext(ctx, query,
		p.Number,
		p.StartTime,
		p.EndTime,
		p.IsDaytime,
		p.Temperature,
		p.TemperatureUnit,
		p.WindSpeed,
		p.WindDirection,
		p.ShortForecast,
		p.GridpointID)

	return err
}

// Update writes this PeriodEntity as an update. The period being updated in the
// database is identified by this PeriodEntity Number and GridpointID fields.
// All fields being updated must be set before calling this method. Number and
// GridpointID cannot be updated.
func (p *PeriodEntity) Update(ctx context.Context, db *sql.DB) error {
	query := `UPDATE periods SET starts = $1, ends = $2, is_day_time = $3, temp = $4,
			  temp_unit = $5, wind_speed = $6, wind_direction = $7, short_forecast = $8
			  WHERE num = $9 AND gp_id = $10`

	_, err := db.ExecContext(ctx, query,
		p.StartTime,
		p.EndTime,
		p.IsDaytime,
		p.Temperature,
		p.TemperatureUnit,
		p.WindSpeed,
		p.WindDirection,
		p.ShortForecast,
		p.Number,
		p.GridpointID)

	return err
}

// PeriodEntityCollection is a collection of PeriodEntity.
type PeriodEntityCollection []PeriodEntity

// ToPeriods returns this PeriodEntityCollection as a PeriodCollection.
func (p *PeriodEntityCollection) ToPeriods() PeriodCollection {
	periods := PeriodCollection{}
	for _, entity := range *p {
		periods = append(periods, entity.ToPeriod())
	}
	return periods
}

// Select reads all the periods in ascending order from the database that
// belong to the specified gridpoint into this PeriodEntityCollection.
func (p *PeriodEntityCollection) Select(ctx context.Context, db *sql.DB, gridpointID int) error {
	query := `SELECT num, starts, ends, is_day_time, temp, temp_unit, wind_speed, 
			  wind_direction, short_forecast, gp_id FROM periods 
			  WHERE gp_id = $1 
			  ORDER BY num`

	rows, err := db.QueryContext(ctx, query, gridpointID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		period := PeriodEntity{}
		if err := period.Scan(rows); err != nil {
			return err
		}
		*p = append(*p, period)
	}

	return nil
}

// Insert writes all the PeriodEntity in this PeriodEntityCollectionn to the
// database. The GridpointID of each PeriodEntity is set to gridpointID before
// being written. All other fields must be set for each PeriodEntity before
// calling this method.
func (p *PeriodEntityCollection) Insert(ctx context.Context, db *sql.DB, gridpointID int) error {
	for i := range *p {
		entity := (*p)[i]
		entity.GridpointID = gridpointID
		if err := entity.Insert(ctx, db); err != nil {
			return err
		}
	}

	return nil
}

// Update writes all the PeriodEntity in this PeriodEntityCollection to the
// database as an update. The GridpointID of each PeriodEntity is set to gridpointID
// before being written. All other fields must be set for each PeriodEntity before
// calling this method.
func (p *PeriodEntityCollection) Update(ctx context.Context, db *sql.DB, gridpointID int) error {
	for i := range *p {
		entity := (*p)[i]
		entity.GridpointID = gridpointID
		if err := entity.Update(ctx, db); err != nil {
			return err
		}
	}

	return nil
}
