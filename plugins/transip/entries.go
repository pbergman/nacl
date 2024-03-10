package main

import "strings"

func GetDnsEntries(config *Config, domains []string) map[string][]*DNSEntry {
	var entries = make(map[string][]*DNSEntry)
	for _, host := range config.Hosts {
		for _, domain := range domains {
			if _, ok := entries[domain]; !ok {
				entries[domain] = make([]*DNSEntry, 0)
			}
			if strings.HasSuffix(host, domain) {
				if host == domain {
					entries[domain] = append(entries[domain], &DNSEntry{
						Type: "A",
						Name: "@",
					})
					if config.Ipv6 {
						entries[domain] = append(entries[domain], &DNSEntry{
							Type: "AAAA",
							Name: "@",
						})
					}
				} else {
					var name = host[0 : len(host)-len(domain)-1]
					entries[domain] = append(entries[domain], &DNSEntry{
						Type: "A",
						Name: name,
					})
					if config.Ipv6 {
						entries[domain] = append(entries[domain], &DNSEntry{
							Type: "AAAA",
							Name: name,
						})
					}
				}
				break
			}
		}
	}
	return entries
}
