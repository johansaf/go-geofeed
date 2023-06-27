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

func getWhoisData(net netip.Prefix, url string) (WhoisResult, error) {
	var whoisResult WhoisResult

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return whoisResult, err
	}

	req.Header.Set("Accept", "application/xml")
	//req.Header.Set("User-Agent", "xxx")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return whoisResult, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return whoisResult, err
	}

	err = xml.Unmarshal(body, &whoisResult)
	if err != nil {
		return whoisResult, err
	}

	return whoisResult, nil
}

func getSupernetData(supernet netip.Prefix) Allocation {
	var allocation Allocation
	var url string

	if supernet.Addr().Is4() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum", supernet)
	} else if supernet.Addr().Is6() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num", supernet)
	}

	whoisResult, err := getWhoisData(supernet, url)
	if err != nil {
		log.Fatal(err)
	}

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
	var url string

	if supernet.Addr().Is4() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inetnum&flags=M", supernet)
	} else if supernet.Addr().Is6() {
		url = fmt.Sprintf("https://rest.db.ripe.net/search?source=ripe&query-string=%s&flags=no-referenced&type-filter=inet6num&flags=M", supernet)
	}

	whoisResult, err := getWhoisData(supernet, url)
	if err != nil {
		log.Fatal(err)
	}

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

	log.Print("Generating geofeed...")

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
