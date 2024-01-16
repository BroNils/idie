package util

import (
	"fmt"
	"net"
)

func GenerateIPRange(startInput, endInput string) ([]string, error) {
	startIP := net.ParseIP(startInput)
	endIP := net.ParseIP(endInput)

	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	if bytesLessThan(endIP, startIP) {
		return nil, fmt.Errorf("end IP address is less than start IP address")
	}

	var ipRange []string

	for ip := startIP; !ip.Equal(endIP); incrementIP(ip) {
		ipRange = append(ipRange, ip.String())
	}

	ipRange = append(ipRange, endIP.String())

	return ipRange, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func bytesLessThan(a, b net.IP) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true
		}
	}

	return false
}
