package server

import (
	"context"
	"log"
	"time"

	"github.com/cicconee/weather-app/internal/alert"
)

type worker struct {
	alerts *alert.Service
	d      time.Duration
	killCh <-chan struct{}
}

func (w *worker) start() {
	ticker := time.NewTicker(w.d)

	for {
		select {
		case <-ticker.C:
			// Execute any jobs.
			ctx := context.Background()
			w.syncAlerts(ctx)
		case <-w.killCh:
			ticker.Stop()
			// TODO: clean up any running jobs.
			return
		}
	}
}

func (w *worker) syncAlerts(ctx context.Context) {
	sync, err := w.alerts.Sync(ctx)
	if err != nil {
		log.Printf("failed syncing alerts: %v\n", err)
	} else {
		for _, fail := range sync.Fails {
			log.Printf("failed to sync alert (id=%s, op=%s): %v\n",
				fail.ID,
				fail.Op,
				fail.Err)
		}

		log.Printf("total alerts written: %d", sync.TotalWrites)
	}

	deleted, err := w.alerts.CleanUp(ctx)
	if err != nil {
		log.Printf("failed to delete outdated alerts: %v\n", err)
	}

	log.Printf("total deletes: %d\n", deleted)
}
