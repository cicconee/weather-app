package state

import (
	"context"

	"github.com/cicconee/weather-app/internal/nws"
	"github.com/cicconee/weather-app/internal/pool"
)

type FetchFailure struct {
	URI string `json:"uri"`
	err error
}

type Fetcher struct {
	client *nws.Client
	p      *pool.Pool
	dataCh chan Zone
	failCh chan FetchFailure
}

func NewFetcher(c *nws.Client, p *pool.Pool, s *Store, zoneCount int) *Fetcher {
	return &Fetcher{
		client: c,
		p:      p,
		dataCh: make(chan Zone, zoneCount),
		failCh: make(chan FetchFailure, zoneCount),
	}
}

func (f *Fetcher) close() {
	close(f.dataCh)
	close(f.failCh)
}

func (w *Fetcher) fail(z Zone, err error) {
	w.failCh <- FetchFailure{
		URI: z.URI,
		err: err,
	}
}

func (w *Fetcher) finish(z Zone) {
	w.dataCh <- z
}

type FetchResult struct {
	Zones ZoneURIMap
	Fails map[string]error
}

// FetchEach concurrently fetches the data for
// each zone in zones. Each zone in the zones slice
// only needs the Type and Code to be set. The
// FetchResult Zones field will hold the up to date
// data for the Zone.
//
// For each zone in zones, if ID, CreatedAt, or
// UpdatedAt was set it will be included in the
// FetchResult Zones field.
func (f *Fetcher) FetchEach(ctx context.Context, zones []Zone) FetchResult {
	result := FetchResult{
		Zones: ZoneURIMap{},
		Fails: map[string]error{},
	}

	// Fetch zone data from the NWS
	// API concurrently.
	for i := range zones {
		f.Fetch(ctx, zones[i])
	}

	// Write each zone to the Zones map.
	// Write each fail to the FetchFailure
	// slice.
	for range zones {
		select {
		case zone := <-f.dataCh:
			result.Zones[zone.URI] = zone
		case fail := <-f.failCh:
			result.Fails[fail.URI] = fail.err
		}
	}

	return result
}

func (f *Fetcher) Fetch(ctx context.Context, z Zone) {
	f.p.Add(func() {
		// Check if context has already been
		// cancelled or timed out before executing
		// long running task.
		if ctx.Err() != nil {
			f.fail(z, ctx.Err())
			return
		}

		zone, err := f.fetch(ctx, z.Type, z.Code)
		if err != nil {
			f.fail(z, err)
			return
		}

		z.CopyUpdateableData(zone)

		f.finish(z)
	})
}

func (f *Fetcher) fetch(ctx context.Context, zoneType string, zoneCode string) (Zone, error) {
	nwsZone, err := f.client.GetZone(zoneType, zoneCode)
	if err != nil {
		return Zone{}, err
	}

	return zoneFromNWS(nwsZone), nil
}
