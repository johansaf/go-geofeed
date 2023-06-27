package main

import (
	"net/netip"
	"time"
)

// Contains the supernets to be used for geofeed generation
var supernets = []string{"192.0.2.0/24", "2001:db8::/32"}

// Contains the geofeed
var geofeed Geofeed

// Contains the result coming from a whois query
type WhoisResult struct {
	Objects []struct {
		Attributes []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"attributes>attribute"`
	} `xml:"objects>object"`
}

// Contains the geofeed
type Geofeed struct {
	Generated   time.Time
	Allocations []Allocation
}

// Contains the supernet information, and any subnets that are not in the same country
type Allocation struct {
	Prefix  netip.Prefix
	Country string
	Subnets []Subnet
}

// Contains a single subnet
type Subnet struct {
	Prefix  netip.Prefix
	Country string
}
