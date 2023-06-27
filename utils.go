package main

import (
	"encoding/binary"
	"math"
	"net/netip"
	"strings"
)

func calcIPv4Cidr(start, end netip.Addr) netip.Prefix {
	var maskSize int

	// Convert the IP addresses to uint32 and calculate the difference
	diff := binary.BigEndian.Uint32(end.AsSlice()) - binary.BigEndian.Uint32(start.AsSlice())

	// Handle special cases
	if diff == 1 {
		maskSize = 31
	} else if diff == 0 {
		maskSize = 32
	} else {
		// Calculate number of trailing zeroes, 32 is the number of bits in an IPv4 address
		maskSize = 32 - int(math.Ceil(math.Log2(float64(diff))))
	}

	return netip.PrefixFrom(start, maskSize)
}

// Function to take a comma separated list of networks and return a slice of strings
func parseNetworks(networks string) []string {
	return strings.Split(networks, ",")
}
