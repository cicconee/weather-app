package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type LogWriter struct {
	logger *log.Logger
	rw     http.ResponseWriter
	r      *http.Request
}

func NewLogWriter(l *log.Logger, rw http.ResponseWriter, r *http.Request) *LogWriter {
	return &LogWriter{l, rw, r}
}

func (l *LogWriter) log(format string, v ...any) {
	l.logger.Println(fmt.Sprintf(format, v...))
}

func (l *LogWriter) Write(r Response) {
	l.rw.Header().Set("Content-Type", "application/json")
	l.rw.WriteHeader(r.Status)
	if err := json.NewEncoder(l.rw).Encode(r.Body); err != nil {
		l.log("*LogWriter.Write: failed to write json to http.ResponseWriter: %v\n", err)
	}
}

type ServerErrorResponser interface {
	ServerErrorResponse() (int, string)
}

func (w *LogWriter) WriteError(err error) {
	errResp := ErrorResponse{
		Status:   http.StatusInternalServerError,
		ErrorMsg: "Something went wrong",
	}

	var apiError ServerErrorResponser
	if errors.As(err, &apiError) {
		errResp.Status, errResp.ErrorMsg = apiError.ServerErrorResponse()
	}

	w.Write(errResp.AsResponse())
}
