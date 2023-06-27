package main

import (
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"time"

	"github.com/go-co-op/gocron"
)

func generateGeofeed() {
	var feed Geofeed

	log.Print("Generating geofeed...")

	for _, supernet := range supernets {
		tmp, err := netip.ParsePrefix(supernet)
		if err != nil {
			log.Println(err)
			continue
		}

		allocationData, err := getSupernetData(tmp)
		if err != nil {
			log.Println(err)
			continue
		}

		subnetData, err := getSubnetData(tmp)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, i := range subnetData {
			if i.Country != allocationData.Country {
				allocationData.Subnets = append(allocationData.Subnets, i)
			}
		}

		feed.Allocations = append(feed.Allocations, allocationData)
	}

	feed.Generated = time.Now().UTC()
	geofeed = feed

	log.Println("Geofeed generation done")
}

func handleGeofeed(w http.ResponseWriter, r *http.Request) {
	if len(geofeed.Allocations) == 0 {
		// Return a 503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Add a Last-Modified header
	w.Header().Set("Last-Modified", geofeed.Generated.Format(time.RFC1123))

	fmt.Fprintf(w, "# Generated %s\n", geofeed.Generated.Format(time.RFC3339Nano))

	for _, allocation := range geofeed.Allocations {
		fmt.Fprintf(w, "%s,%s,,,\n", allocation.Prefix, allocation.Country)
		for _, subnet := range allocation.Subnets {
			fmt.Fprintf(w, "%s,%s,,,\n", subnet.Prefix, subnet.Country)
		}
	}

	fmt.Fprintf(w, "# EOF\n")
}

func handleRegenerateGeofeed(w http.ResponseWriter, r *http.Request) {
	// TODO: Check if the request is coming from a trusted source, or using a secret token
	// Regenerate the geofeed
	generateGeofeed()
	// Return a 200 OK
	w.WriteHeader(http.StatusOK)
}

func main() {

	sched := gocron.NewScheduler(time.UTC)
	sched.EveryRandom(24, 36).Hours().Do(generateGeofeed)
	sched.StartAsync()

	log.Println("Starting server...")
	http.HandleFunc("/", handleGeofeed)
	http.HandleFunc("/generate", handleRegenerateGeofeed)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
