package main

import (
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"time"
)

// Generates the geofeed and stores it into memory
func generateGeofeed() {
	var feed Geofeed

	log.Print("Generating geofeed...")

	for _, supernet := range cfg.Networks {
		tmp, err := netip.ParsePrefix(supernet)
		if err != nil {
			log.Println(err)
			continue
		}

		// Get the supernet data
		allocationData, err := getSupernetData(tmp)
		if err != nil {
			log.Println(err)
			continue
		}

		// Get all subnets contained within the supernet
		subnetData, err := getSubnetData(tmp)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, i := range subnetData {
			// If the subnet country is different from the supernet country then add it to the list
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
		// If the geofeed is still being generated we return a 503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Last-Modified", geofeed.Generated.Format(time.RFC1123))

	fmt.Fprintf(w, "# Generated %s\n", geofeed.Generated.Format(time.RFC3339Nano))

	// Loop through the geofeed struct and print out the entries
	for _, allocation := range geofeed.Allocations {
		fmt.Fprintf(w, "%s,%s,,,\n", allocation.Prefix, allocation.Country)
		for _, subnet := range allocation.Subnets {
			fmt.Fprintf(w, "%s,%s,,,\n", subnet.Prefix, subnet.Country)
		}
	}

	fmt.Fprintf(w, "# EOF\n")
}

// Regenerates the geofeed if we want to force an update
func handleRegenerateGeofeed(w http.ResponseWriter, r *http.Request) {
	if cfg.Key == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Header.Get("X-Geofeed-Key") != cfg.Key {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	generateGeofeed()
	w.WriteHeader(http.StatusOK)
}
