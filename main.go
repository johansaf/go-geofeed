package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
)

func main() {
	sched := gocron.NewScheduler(time.UTC)
	sched.EveryRandom(24, 36).Hours().Do(generateGeofeed)
	sched.StartAsync()

	log.Println("Starting server...")
	http.HandleFunc("/", handleGeofeed)
	http.HandleFunc("/generate", handleRegenerateGeofeed)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
