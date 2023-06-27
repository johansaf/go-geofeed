package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var cfg = Config{}

// Function to take a comma separated list of networks and return a slice of strings
func parseNetworks(networks string) []string {
	return strings.Split(networks, ",")
}

func readConfig() Config {
	// Check if the LISTEN_ADDRESS environment variable is set, set to ":8080" if not
	if os.Getenv("LISTEN_ADDRESS") == "" {
		os.Setenv("LISTEN_ADDRESS", ":8080")
	}

	// Check if REFRESH_INTERVAL variables
	if os.Getenv("REFRESH_INTERVAL_MIN") == "" {
		os.Setenv("REFRESH_INTERVAL_MIN", "24")
	}

	if os.Getenv("REFRESH_INTERVAL_MAX") == "" {
		os.Setenv("REFRESH_INTERVAL_MAX", "36")
	}

	// Split the list of networks into a slice
	if os.Getenv("NETWORKS") == "" {
		log.Fatal("NETWORKS environment variable not set")
	}

	networks := parseNetworks(os.Getenv("NETWORKS"))

	cfg := Config{
		ListenAddress:      os.Getenv("LISTEN_ADDRESS"),
		RefreshIntervalMin: 24,
		RefreshIntervalMax: 36,
		Networks:           networks,
	}

	return cfg
}

func main() {
	cfg = readConfig()

	sched := gocron.NewScheduler(time.UTC)
	sched.EveryRandom(cfg.RefreshIntervalMin, cfg.RefreshIntervalMax).Hours().Do(generateGeofeed)
	sched.StartAsync()

	log.Printf("Listening on %s", cfg.ListenAddress)
	log.Printf("Refresh interval: %d-%d hours", cfg.RefreshIntervalMin, cfg.RefreshIntervalMax)
	http.HandleFunc("/", handleGeofeed)
	http.HandleFunc("/generate", handleRegenerateGeofeed)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, nil))
}
