package server

import (
	"bytes"
	"log"
	"net/http"
)

type Handler struct {
	logger *log.Logger
}

func NewHandler(l *log.Logger) *Handler {
	return &Handler{
		logger: l,
	}
}

func (h *Handler) HelloWorld() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes.NewBufferString(`{"message": "Hello, World!"}`).Bytes())
	}
}
