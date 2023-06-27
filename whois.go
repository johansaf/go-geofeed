package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/netip"
	"strings"
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
