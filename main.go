package main

import (
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"time"

	"github.com/go-co-op/gocron"
)

func getSupernetData(supernet netip.Prefix) (Allocation, error) {
	var allocation Allocation
	var url string

	if supernet.Addr().Is4() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum", supernet)
	} else if supernet.Addr().Is6() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num", supernet)
	}

	whoisResult, err := getWhoisData(supernet, url)
	if err != nil {
		return allocation, err
	}

	data, err := parseWhoisResult(whoisResult)
	if err != nil {
		return allocation, err
	}
	allocation.Prefix = data[0].Prefix
	allocation.Country = data[0].Country

	return allocation, err
}

func getSubnetData(supernet netip.Prefix) ([]Subnet, error) {
	var subnets []Subnet
	var url string

	if supernet.Addr().Is4() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum&flags=M", supernet)
	} else if supernet.Addr().Is6() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num&flags=M", supernet)
	}

	whoisResult, err := getWhoisData(supernet, url)
	if err != nil {
		return subnets, err
	}

	subnets, err = parseWhoisResult(whoisResult)
	if err != nil {
		return subnets, err
	}

	return subnets, err
}

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
