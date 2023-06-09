package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cicconee/weather-app/internal/admin"
	"github.com/cicconee/weather-app/internal/alert"
	"github.com/cicconee/weather-app/internal/forecast"
	"github.com/cicconee/weather-app/internal/nws"
	"github.com/cicconee/weather-app/internal/pool"
	"github.com/cicconee/weather-app/internal/server"
	"github.com/cicconee/weather-app/internal/state"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var port string

// TODO: Make secretKey a environment variable.
var secretKey = "secret-key"

func main() {
	flag.StringVar(&port, "p", "8080", "the port the server should listen on")
	flag.Parse()

	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "weather_app", "password", "0.0.0.0", "5432", "weather_app_db")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a Pool with 10 workers each
	// with a channel size of 100.
	pool := pool.New(10, 100)
	pool.Start()

	srv := server.Server{
		Addr:      port,
		Router:    chi.NewRouter(),
		Interval:  10 * time.Second,
		Logger:    log.Default(),
		States:    state.New(nws.DefaultClient, db, pool),
		Alerts:    alert.New(nws.DefaultClient, db),
		Forecasts: forecast.New(nws.DefaultClient, db),
		Admins:    admin.New([]byte(secretKey), db),
	}
	if err := srv.Start(); err != nil {
		log.Println(err)
	}
}
