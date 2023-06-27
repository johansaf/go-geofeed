package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/netip"
	"strings"
)

// Makes a GET request to the whois database API and returns the unmashalled result
func getWhoisData(net netip.Prefix, url string) (WhoisResult, error) {
	var whoisResult WhoisResult

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return whoisResult, err
	}

	req.Header.Set("Accept", "application/xml")
	req.Header.Set("User-Agent", "go-geofeed/contact "+cfg.Email)

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

// Loops through the whois result and returns a slice of subnets
func parseWhoisResult(whoisResult WhoisResult) ([]Subnet, error) {
	var subnets []Subnet

	for _, object := range whoisResult.Objects {
		var subnet Subnet
		for _, attribute := range object.Attributes {
			if attribute.Name == "inetnum" {
				// The inetnum attribute contains the start and end IP address of the subnet
				// in the format of "192.0.2.0 - 192.0.2.255", but we need it in CIDR notation
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
				// inet6num always uses CIDR notation
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
