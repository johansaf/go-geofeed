package main

import (
	"net/netip"
	"time"
)

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
