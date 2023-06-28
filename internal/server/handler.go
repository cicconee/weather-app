package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cicconee/weather-app/internal/alert"
	"github.com/cicconee/weather-app/internal/geometry"
	"github.com/cicconee/weather-app/internal/state"
)

type Handler struct {
	logger *log.Logger
	states *state.Service
	alerts *alert.Service
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

func (h *Handler) HandleSyncState() http.HandlerFunc {
	type res struct {
		State        string                  `json:"state"`
		TotalInserts int                     `json:"total_inserts"`
		TotalUpdates int                     `json:"total_updates"`
		TotalDeletes int                     `json:"total_deletes"`
		Fails        []state.SyncZoneFailure `json:"fails"`
		UpdatedAt    time.Time               `json:"created_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		stateID := r.URL.Query().Get("q")
		ctx := r.Context()
		writer := h.NewLogWriter(w, r)

		result, err := h.states.Sync(ctx, stateID)
		if err != nil {
			h.logger.Printf("HandlerSyncState: failed to sync state (stateID=%q): %v", stateID, err)
			writer.WriteError(err)
			return
		}

		writer.Write(Response{
			Status: http.StatusOK,
			Body: res{
				State:        result.State,
				TotalInserts: len(result.Inserts),
				TotalUpdates: len(result.Updates),
				TotalDeletes: len(result.Deletes),
				Fails:        result.Fails,
			},
		})
	}
}

func (h *Handler) HandleGetAlerts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lon := r.URL.Query().Get("lon")
		lat := r.URL.Query().Get("lat")
		writer := h.NewLogWriter(w, r)

		point, err := extractPoint(lon, lat)
		if err != nil {
			h.logger.Printf("HandleGetAlerts: failed to extract point (lon=%q, lat=%q): %v", lon, lat, err)
			writer.WriteError(err)
			return
		}

		getResponse, err := h.alerts.Get(ctx, point)
		if err != nil {
			h.logger.Printf("HandleGetAlerts: failed to get alerts (point=%v): %v", point, err)
			writer.WriteError(err)
			return
		}

		writer.Write(Response{
			Status: http.StatusOK,
			Body:   getResponse,
		})
	}
}

func extractPoint(lonStr string, latStr string) (geometry.Point, error) {
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return geometry.Point{}, fmt.Errorf("failed to parse lon: %w", err)
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return geometry.Point{}, fmt.Errorf("failed to parse lat: %w", err)
	}

	return geometry.NewPoint(lon, lat), nil
}
