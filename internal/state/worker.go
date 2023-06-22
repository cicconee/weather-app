package state

import (
	"context"

	"github.com/cicconee/weather-app/internal/nws"
	"github.com/cicconee/weather-app/internal/pool"
)

type worker struct {
	client *nws.Client
	p      *pool.Pool
	dataCh chan Zone
	failCh chan SaveZoneFailure
}

func newWorker(c *nws.Client, p *pool.Pool, zoneCount int) *worker {
	return &worker{
		client: c,
		p:      p,
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
		SaveZoneResult: SaveZoneResult{
			URI:  z.URI,
			Code: z.Code,
			Type: z.Type,
		},
		err: err,
	}
}

func (w *worker) finish(z Zone) {
	w.dataCh <- z
}

func (w *worker) FetchEach(ctx context.Context, zones []Zone) {
	for i := range zones {
		w.Fetch(ctx, zones[i])
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

		z.Geometry = zone.Geometry

		w.finish(z)
	})
}
