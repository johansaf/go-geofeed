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
