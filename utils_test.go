package main

import (
	"net/netip"
	"testing"
)

func TestCalcIPv4Cidr(t *testing.T) {
	var start, end netip.Addr
	var cidr netip.Prefix

	// /24
	start = netip.MustParseAddr("192.0.2.0")
	end = netip.MustParseAddr("192.0.2.255")
	cidr = calcIPv4Cidr(start, end)
	if cidr.Bits() != 24 {
		t.Errorf("Expected 24, got %s", cidr.String())
	}

	// /25
	start = netip.MustParseAddr("192.0.2.0")
	end = netip.MustParseAddr("192.0.2.127")
	cidr = calcIPv4Cidr(start, end)
	if cidr.Bits() != 25 {
		t.Errorf("Expected 25, got %s", cidr.String())
	}

	// /30
	start = netip.MustParseAddr("192.0.2.0")
	end = netip.MustParseAddr("192.0.2.3")
	cidr = calcIPv4Cidr(start, end)
	if cidr.Bits() != 30 {
		t.Errorf("Expected 30, got %s", cidr.String())
	}

	// /31
	start = netip.MustParseAddr("192.0.2.0")
	end = netip.MustParseAddr("192.0.2.1")
	cidr = calcIPv4Cidr(start, end)
	if cidr.Bits() != 31 {
		t.Errorf("Expected 31, got %s", cidr.String())
	}

	// /32
	start = netip.MustParseAddr("192.0.2.0")
	end = netip.MustParseAddr("192.0.2.0")
	cidr = calcIPv4Cidr(start, end)
	if cidr.Bits() != 32 {
		t.Errorf("Expected 32, got %s", cidr.String())
	}
}

func TestParseNetworks(t *testing.T) {
	networks := parseNetworks("192.0.2.0/24,2001:db8::/32")
	if len(networks) != 2 {
		t.Errorf("Expected 2, got %d", len(networks))
	}
	if networks[0] != "192.0.2.0/24" {
		t.Errorf("Expected 192.0.2.0/24, got %s", networks[0])
	}
	if networks[1] != "2001:db8::/32" {
		t.Errorf("Expected 2001:db8::/32, got %s", networks[1])
	}
}
