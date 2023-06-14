package server

import (
	"fmt"
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
			fmt.Println("at interval.")
		case <-w.killCh:
			fmt.Println("killing interval...")
			ticker.Stop()
			time.Sleep(3 * time.Second)
			fmt.Println("interval killed.")
			return
		}
	}
}
