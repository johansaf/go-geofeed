package main

import (
	"encoding/binary"
	"log"
	"math"
	"net/netip"
	"os"
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

func readConfig() Config {
	// Check if the LISTEN_ADDRESS environment variable is set, set to ":8080" if not
	if os.Getenv("LISTEN_ADDRESS") == "" {
		os.Setenv("LISTEN_ADDRESS", ":8080")
	}

	// Check if REFRESH_INTERVAL variables
	if os.Getenv("REFRESH_INTERVAL_MIN") == "" {
		os.Setenv("REFRESH_INTERVAL_MIN", "24")
	}

	if os.Getenv("REFRESH_INTERVAL_MAX") == "" {
		os.Setenv("REFRESH_INTERVAL_MAX", "36")
	}

	// Split the list of networks into a slice
	if os.Getenv("NETWORKS") == "" {
		log.Fatal("NETWORKS environment variable not set")
	}

	networks := parseNetworks(os.Getenv("NETWORKS"))

	// Require an e-mail address to be set, incase the database operator needs to contact you
	if os.Getenv("EMAIL") == "" {
		log.Fatal("EMAIL environment variable not set")
	}

	email := os.Getenv("EMAIL")

	cfg := Config{
		ListenAddress:      os.Getenv("LISTEN_ADDRESS"),
		RefreshIntervalMin: 24,
		RefreshIntervalMax: 36,
		Networks:           networks,
		Email:              email,
		Key:                os.Getenv("KEY"),
	}

	return cfg
}
