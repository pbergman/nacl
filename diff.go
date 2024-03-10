package main

import (
	"net"
)

func diff(a, b []net.Addr, control net.IPNet) (net.IP, net.IP) {
	var listB = make(map[string]net.IP, len(b))
	var listA = make(map[string]net.IP, len(b))
	var ipv = 1

	var diff = [2]net.IP{
		nil,
		control.IP,
	}

	if ip := control.IP.To4(); ip != nil && len(ip) == net.IPv4len {
		ipv = 0
	}

	for _, x := range a {
		var ip = x.(*net.IPNet).IP

		switch ipv {
		case 0:
			if ipv4 := ip.To4(); ipv4 != nil && len(ipv4) == net.IPv4len {
				listA[ipv4.String()] = ipv4
			}
		case 1:
			if ipv6 := ip.To16(); ipv6 != nil && len(ipv6) == net.IPv6len {
				listA[ipv6.String()] = ipv6
			}
		}
	}

	for _, x := range b {
		var ip = x.(*net.IPNet).IP

		switch ipv {
		case 0:
			if ipv4 := ip.To4(); ipv4 != nil && len(ipv4) == net.IPv4len {
				listB[ipv4.String()] = ipv4
			}
		case 1:
			if ipv6 := ip.To16(); ipv6 != nil && len(ipv6) == net.IPv6len {
				listB[ipv6.String()] = ipv6
			}
		}
	}

	for idx, ip := range listA {
		if _, found := listB[idx]; !found {

			switch ipv {
			case 0:
				if ipv4 := ip.To4(); ipv4 != nil && len(ipv4) == net.IPv4len {
					diff[0] = ipv4
				}
			case 1:
				if ipv6 := ip.To16(); ipv6 != nil && len(ipv6) == net.IPv6len {
					diff[0] = ipv6
				}
			}

		}
	}

	return diff[0], diff[1]
}
