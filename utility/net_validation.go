package utility

import (
	"fmt"
	"net"
)

var blockedNetworks []*net.IPNet

func init() {
	cidrs := []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918
		"192.168.0.0/16", // RFC 1918
		"169.254.0.0/16", // link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local
	}
	for _, cidr := range cidrs {
		_, network, _ := net.ParseCIDR(cidr)
		blockedNetworks = append(blockedNetworks, network)
	}
}

func isBlockedIP(ip net.IP) bool {
	for _, network := range blockedNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateExternalHost checks that a hostname resolves to a public (non-internal) IP address.
func ValidateExternalHost(host string) error {
	if len(host) == 0 {
		return fmt.Errorf("host must not be empty")
	}

	// If the host is a raw IP, check it directly
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("host %s resolves to a blocked internal address", host)
		}
		return nil
	}

	// Resolve hostname and check all IPs
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %s: %w", host, err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip != nil && isBlockedIP(ip) {
			return fmt.Errorf("host %s resolves to a blocked internal address (%s)", host, ipStr)
		}
	}
	return nil
}
