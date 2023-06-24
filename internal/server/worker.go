package server

import (
	"time"
)

type worker struct {
	d      time.Duration
	killCh <-chan struct{}
}

func (w *worker) start() {
	ticker := time.NewTicker(w.d)

	for {
		select {
		case <-ticker.C:
			// TODO: execute job.
		case <-w.killCh:
			ticker.Stop()
			// TODO: clean up any running jobs.
			return
		}
	}
}
