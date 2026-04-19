package handler

import (
	"testing"
)

func TestIsIPInCIDR(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		cidr      string
		expected  bool
		wantError bool
	}{
		{
			name:      "IP in CIDR range",
			ip:        "192.168.1.10",
			cidr:      "192.168.1.0/24",
			expected:  true,
			wantError: false,
		},
		{
			name:      "IP not in CIDR range",
			ip:        "192.168.2.10",
			cidr:      "192.168.1.0/24",
			expected:  false,
			wantError: false,
		},
		{
			name:      "Loopback in loopback range",
			ip:        "127.0.0.1",
			cidr:      "127.0.0.0/8",
			expected:  true,
			wantError: false,
		},
		{
			name:      "Valid single IP as /32",
			ip:        "10.0.0.1",
			cidr:      "10.0.0.1/32",
			expected:  true,
			wantError: false,
		},
		{
			name:      "Different single IP",
			ip:        "10.0.0.2",
			cidr:      "10.0.0.1/32",
			expected:  false,
			wantError: false,
		},
		{
			name:      "Invalid IP address",
			ip:        "not-an-ip",
			cidr:      "192.168.1.0/24",
			expected:  false,
			wantError: true,
		},
		{
			name:      "Invalid CIDR notation",
			ip:        "192.168.1.10",
			cidr:      "192.168.1.0/33",
			expected:  false,
			wantError: true,
		},
		{
			name:      "Empty CIDR",
			ip:        "192.168.1.10",
			cidr:      "",
			expected:  false,
			wantError: true,
		},
		{
			name:      "IPv6 in IPv6 range",
			ip:        "2001:db8::1",
			cidr:      "2001:db8::/32",
			expected:  true,
			wantError: false,
		},
		{
			name:      "IPv6 not in IPv6 range",
			ip:        "2001:db9::1",
			cidr:      "2001:db8::/32",
			expected:  false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsIPInCIDR(tt.ip, tt.cidr)

			if (err != nil) != tt.wantError {
				t.Errorf("IsIPInCIDR() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if result != tt.expected {
				t.Errorf("IsIPInCIDR() = %v, want %v", result, tt.expected)
			}
		})
	}
}
