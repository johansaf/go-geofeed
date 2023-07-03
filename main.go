package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
)

var cfg = Config{}

func main() {
	cfg = readConfig()

	sched := gocron.NewScheduler(time.UTC)
	sched.EveryRandom(cfg.RefreshIntervalMin, cfg.RefreshIntervalMax).Hours().Do(generateGeofeed)
	sched.StartAsync()

	log.Printf("Listening on %s", cfg.ListenAddress)
	log.Printf("Refresh interval: %d-%d hours", cfg.RefreshIntervalMin, cfg.RefreshIntervalMax)
	http.HandleFunc("/", handleUnknown)
	http.HandleFunc("/geofeed.csv", handleGeofeed)
	http.HandleFunc("/generate", handleRegenerateGeofeed)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, nil))
}
