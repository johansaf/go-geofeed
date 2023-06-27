package main

import (
	"net/netip"
	"time"
)

var supernets = []string{"192.0.2.0/24", "2001:db8::/32"}
var geofeed Geofeed

type WhoisResult struct {
	Objects []struct {
		Attributes []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"attributes>attribute"`
	} `xml:"objects>object"`
}

type Geofeed struct {
	Generated   time.Time
	Allocations []Allocation
}

type Subnet struct {
	Prefix  netip.Prefix
	Country string
}

type Allocation struct {
	Prefix  netip.Prefix
	Country string
	Subnets []Subnet
}
