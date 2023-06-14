package server

import (
	"bytes"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
}

func (h *Handler) setRoutes(r *chi.Mux) {
	r.Get("/", h.HelloWorld())
}

func (h *Handler) HelloWorld() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes.NewBufferString(`{"message": "Hello, World!"}`).Bytes())
	}
}
