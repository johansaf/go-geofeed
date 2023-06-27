package main

import (
	"fmt"
	"net/netip"
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
