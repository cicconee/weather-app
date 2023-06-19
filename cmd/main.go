package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cicconee/weather-app/internal/server"
	"github.com/cicconee/weather-app/internal/state"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var port string

func main() {
	flag.StringVar(&port, "p", "8080", "the port the server should listen on")
	flag.Parse()

	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "weather_app", "password", "0.0.0.0", "5432", "weather_app_db")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}

	srv := server.Server{
		Addr:     port,
		Router:   chi.NewRouter(),
		Interval: time.Second,
		Logger:   log.Default(),
		States:   state.New(db),
	}
	if err := srv.Start(); err != nil {
		log.Println(err)
	}
}
