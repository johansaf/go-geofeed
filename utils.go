package main

import (
	"encoding/binary"
	"math"
	"net/netip"
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
