package server

import (
	"bytes"
	"net/http"
)

type Handler struct {
}

func (h *Handler) HelloWorld() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes.NewBufferString(`{"message": "Hello, World!"}`).Bytes())
	}
}
