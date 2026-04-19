package handler

import (
	"fmt"
	"net"
)

func IsIPInCIDR(ipStr string, cidrStr string) (bool, error) {
	if cidrStr == "" {
		return false, fmt.Errorf("CIDR subnet is empty")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR notation: %s, error: %w", cidrStr, err)
	}

	return cidr.Contains(ip), nil
}
