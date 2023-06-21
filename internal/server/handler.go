package server

import (
	"log"
	"net/http"
	"time"

	"github.com/cicconee/weather-app/internal/state"
)

type Handler struct {
	logger *log.Logger
	states *state.Service
}

func NewHandler(l *log.Logger) *Handler {
	return &Handler{
		logger: l,
	}
}

func (h *Handler) NewLogWriter(w http.ResponseWriter, r *http.Request) *LogWriter {
	return NewLogWriter(h.logger, w, r)
}

func (h *Handler) HelloWorld() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type res struct {
			Message string `json:"message"`
		}

		h.NewLogWriter(w, r).Write(Response{
			Status: http.StatusOK,
			Body:   res{Message: "Hello, World!"},
		})
	}
}

func (h *Handler) HandleCreateState() http.HandlerFunc {
	type res struct {
		State       string                  `json:"state"`
		TotalZones  int                     `json:"total_zones"`
		TotalWrites int                     `json:"total_writes"`
		Fails       []state.SaveZoneFailure `json:"fails"`
		CreatedAt   time.Time               `json:"created_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		stateID := r.URL.Query().Get("q")
		ctx := r.Context()
		writer := h.NewLogWriter(w, r)

		result, err := h.states.Save(ctx, stateID)
		if err != nil {
			h.logger.Printf("HandleCreateState: failed to save state (stateID=%q): %v", stateID, err)
			writer.WriteError(err)
			return
		}

		writer.Write(Response{
			Status: http.StatusOK,
			Body: res{
				State:       result.State,
				TotalZones:  result.TotalZones(),
				TotalWrites: len(result.Writes),
				Fails:       result.Fails,
				CreatedAt:   result.CreatedAt,
			},
		})
	}
}
