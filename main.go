package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var supernets = []string{"192.0.2.0/24", "2001:db8::/32"}

var geofeed Geofeed

func getSupernetData(supernet netip.Prefix) Allocation {
	var allocation Allocation
	var base_url string

	if supernet.Addr().Is4() {
		base_url = "https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum"
	} else {
		base_url = "https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num"
	}
	url := fmt.Sprintf(base_url, supernet)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var whoisResult WhoisResult
	xml.Unmarshal(body, &whoisResult)

	// Loop through the objects and look for the inetnum and country attributes, print them when found
	for _, object := range whoisResult.Objects {
		for _, attribute := range object.Attributes {
			if attribute.Name == "inetnum" {
				// Split the inetnum
				inetnum := strings.Split(attribute.Value, " - ")
				// Calculate the CIDR
				x, _ := netip.ParseAddr(inetnum[0])
				y, _ := netip.ParseAddr(inetnum[1])
				allocation.Prefix = calcIPv4Cidr(x, y)
			} else if attribute.Name == "inet6num" {
				allocation.Prefix = netip.MustParsePrefix(attribute.Value)
			}

			if attribute.Name == "country" {
				allocation.Country = attribute.Value
			}
		}
	}

	return allocation
}

func getSubnetData(supernet netip.Prefix) []Subnet {
	var subnets []Subnet
	var base_url string

	if supernet.Addr().Is4() {
		base_url = "https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum&flags=M"
	} else if supernet.Addr().Is6() {
		base_url = "https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num&flags=M"
	}
	url := fmt.Sprintf(base_url, supernet)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var whoisResult WhoisResult
	xml.Unmarshal(body, &whoisResult)

	// Loop through the objects, create a temporary Subnet struct, look for the inetnum and country attributes, append them to the subnets slice
	for _, object := range whoisResult.Objects {
		var subnet Subnet
		for _, attribute := range object.Attributes {
			if attribute.Name == "inetnum" {
				// Split the inetnum
				inetnum := strings.Split(attribute.Value, " - ")
				// Calculate the CIDR
				x, _ := netip.ParseAddr(inetnum[0])
				y, _ := netip.ParseAddr(inetnum[1])
				subnet.Prefix = calcIPv4Cidr(x, y)
			} else if attribute.Name == "inet6num" {
				subnet.Prefix = netip.MustParsePrefix(attribute.Value)
			}
			if attribute.Name == "country" {
				subnet.Country = attribute.Value
			}
		}
		subnets = append(subnets, subnet)
	}

	return subnets
}

func generateGeofeed() {
	var feed Geofeed

	fmt.Print("Generating geofeed... ")

	for _, supernet := range supernets {
		tmp, _ := netip.ParsePrefix(supernet)
		allocationData := getSupernetData(tmp)
		subnetData := getSubnetData(tmp)

		for _, i := range subnetData {
			if i.Country != allocationData.Country {
				allocationData.Subnets = append(allocationData.Subnets, i)
			}
		}

		feed.Allocations = append(feed.Allocations, allocationData)
	}

	feed.Generated = time.Now().UTC()
	geofeed = feed

	fmt.Println("Done")
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

	fmt.Println("Starting server...")
	http.HandleFunc("/", handleGeofeed)
	http.HandleFunc("/generate", handleRegenerateGeofeed)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
