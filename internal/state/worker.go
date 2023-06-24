package state

import (
	"context"

	"github.com/cicconee/weather-app/internal/nws"
	"github.com/cicconee/weather-app/internal/pool"
)

type worker struct {
	client *nws.Client
	p      *pool.Pool
	s      *Store
	dataCh chan Zone
	failCh chan SaveZoneFailure
}

func newWorker(c *nws.Client, p *pool.Pool, s *Store, zoneCount int) *worker {
	return &worker{
		client: c,
		p:      p,
		s:      s,
		dataCh: make(chan Zone, zoneCount),
		failCh: make(chan SaveZoneFailure, zoneCount),
	}
}

func (w *worker) close() {
	close(w.dataCh)
	close(w.failCh)
}

func (w *worker) fail(z Zone, err error) {
	w.failCh <- SaveZoneFailure{
		URI:  z.URI,
		Code: z.Code,
		Type: z.Type,
		err:  err,
	}
}

func (w *worker) finish(z Zone) {
	w.dataCh <- z
}

func (w *worker) SaveEach(ctx context.Context, zones []Zone) SaveZoneResult {
	// Fetch zone data from the NWS
	// API concurrently.
	for i := range zones {
		w.Fetch(ctx, zones[i])
	}

	// Define slices that will hold
	// the write results.
	writes := []Zone{}
	fails := []SaveZoneFailure{}

	// Write each successfully fetched
	// zone to the database. If any
	// errors occurred record it in
	// the fails slice.
	for range zones {
		select {
		case zone := <-w.dataCh:
			if err := w.s.InsertZoneTx(ctx, zone); err != nil {
				fails = append(fails, zone.SaveZoneFailure(err))
			} else {
				writes = append(writes, zone)
			}
		case fail := <-w.failCh:
			fails = append(fails, fail)
		}
	}

	return SaveZoneResult{
		Writes: writes,
		Fails:  fails,
	}
}

func (w *worker) Fetch(ctx context.Context, z Zone) {
	w.p.Add(func() {
		// Check if context has already been
		// cancelled or timed out before executing
		// long running task.
		if ctx.Err() != nil {
			w.fail(z, ctx.Err())
			return
		}

		zone, err := w.client.GetZone(z.Type, z.Code)
		if err != nil {
			w.fail(z, err)
			return
		}

		z.Geometry = NewGeometry(zone.Geometry)

		w.finish(z)
	})
}
