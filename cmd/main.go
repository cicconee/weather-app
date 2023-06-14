package main

import (
	"flag"
	"log"
	"time"

	"github.com/cicconee/weather-app/internal/server"
	"github.com/go-chi/chi/v5"
)

var port string

func main() {
	flag.StringVar(&port, "p", "8080", "the port the server should listen on")
	flag.Parse()

	srv := server.Server{
		Addr:     port,
		Router:   chi.NewRouter(),
		Interval: time.Second,
	}
	if err := srv.Start(); err != nil {
		log.Println(err)
	}
}
