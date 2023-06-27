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

func parseWhoisResult(whoisResult WhoisResult) ([]Subnet, error) {
	var subnets []Subnet

	for _, object := range whoisResult.Objects {
		var subnet Subnet
		for _, attribute := range object.Attributes {
			if attribute.Name == "inetnum" {
				inetnum := strings.Split(attribute.Value, " - ")
				start, err := netip.ParseAddr(inetnum[0])
				if err != nil {
					return subnets, err
				}
				end, err := netip.ParseAddr(inetnum[1])
				if err != nil {
					return subnets, err
				}
				subnet.Prefix = calcIPv4Cidr(start, end)
			} else if attribute.Name == "inet6num" {
				subnet.Prefix = netip.MustParsePrefix(attribute.Value)
			}

			if attribute.Name == "country" {
				subnet.Country = attribute.Value
			}
		}
		subnets = append(subnets, subnet)
	}

	return subnets, nil
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
