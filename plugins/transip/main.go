package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"net"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"github.com/pbergman/nacl/plugin"
	"golang.org/x/sys/unix"
)

var (
	Version string
	Name    plugin.PluginName    = "transip"
	Factory plugin.PluginFactory = func(config toml.Primitive, meta *toml.MetaData, logger *logger.Logger) (plugin.Plugin, error) {

		cnf, err := getConfig(config, meta)

		if err != nil {
			return nil, err
		}

		api, err := NewApiClient(cnf, logger)

		domains, err := api.GetDomains()

		if err != nil {
			return nil, fmt.Errorf("failed to fetch domains from account: %s", err)
		}

		if len(domains) == 0 {
			return nil, fmt.Errorf("no domains registered on transip")
		}

		var records = GetDnsEntries(cnf, domains)

		if 0 == len(records) {
			return nil, fmt.Errorf("no valid domains in config")
		}

		if err != nil {
			return nil, err
		}
		spew.Dump(records)
		return &TransIp{config: cnf, logger: logger, client: api, once: new(sync.Once), records: records}, nil
	}
)

type TransIp struct {
	config  *Config
	client  *ApiClient
	records map[string][]*DNSEntry
	logger  *logger.Logger
	once    *sync.Once
}

func (t *TransIp) boot() {
	for domain, entries := range t.records {
		existing, err := t.client.GetDNSEntries(domain)
		if err != nil {
			// not found so we should create entries with init
			if v, ok := err.(*ApiError); ok && v.Code == 404 {
				continue
			}
			t.logger.Error(fmt.Sprintf("could not fetch dns records: %s", err.Error()))
			continue
		}
		for _, entry := range entries {
			for _, x := range existing {
				if x.Name == entry.Name && x.Type == entry.Type {
					entry.Content = x.Content
					entry.Expire = x.Expire
				}
			}
		}
		for _, entry := range entries {
			for _, x := range existing {
				if x.Name == entry.Name && x.Type == entry.Type {
					entry.Content = x.Content
					entry.Expire = x.Expire
				}
			}
		}
	}
	for domain, entries := range t.records {
		for _, entry := range entries {
			t.logger.Debug(fmt.Sprintf("monitor dns record %s.%s (type: %s, expire: %d, content: %s)", entry.Name, domain, entry.Type, entry.Expire, entry.Content))
		}
	}
}

func (t *TransIp) Init(ip *net.IPNet) error {
	t.once.Do(t.boot)
	for domain, entries := range t.records {
		for _, entry := range entries {
			var shouldUpdate = false
			if ipv4 := ip.IP.To4(); ipv4 != nil && len(ipv4) == net.IPv4len && entry.Type == "A" && entry.Content != ipv4.String() {
				t.logger.Debug(fmt.Sprintf("update %s.%s A record from %s to %s", entry.Name, domain, entry.Content, ipv4.String()))
				entry.Content = ipv4.String()
				shouldUpdate = true
			}
			if ipv6 := ip.IP.To16(); ipv6 != nil && len(ipv6) == net.IPv6len && entry.Type == "AAAA" && entry.Content != ipv6.String() {
				t.logger.Debug(fmt.Sprintf("update %s.%s AAAA record from %s to %s", entry.Name, domain, entry.Content, ipv6.String()))
				entry.Content = ipv6.String()
				shouldUpdate = true
			}
			if shouldUpdate {
				resp, err := t.client.UpdateDNSEntry(domain, entry)
				if err != nil {
					return err
				}
				entry.Content = resp.Content
			}
		}
	}
	return nil
}

func (t *TransIp) Handle(message *netlink.NetlinkMessageContext) error {
	switch message.GetMessageType() {
	case unix.RTM_NEWADDR:
		for domain, entries := range t.records {
			for _, entry := range entries {
				var update *DNSEntry
				switch entry.Type {
				case "A":
					if message.GetAddrFamily() == netlink.IPV4 {
						update = entry
						update.Content = message.GetIpNet().IP.To4().String()
					}
				case "AAAA":
					if message.GetAddrFamily() == netlink.IPV6 {
						update = entry
						update.Content = message.GetIpNet().IP.To16().String()
					}
				}
				if nil != update {

					if ret, err := t.client.UpdateDNSEntry(domain, update); err != nil {
						return err
					} else {
						t.logger.Debug(fmt.Sprintf("update %s.%s %s record from %s to %s", update.Name, domain, update.Type, entry.Content, ret.Content))
						update.Content = ret.Content
					}
				}
			}
		}
	case unix.RTM_DELADDR:
		for domain, entries := range t.records {
			for _, entry := range entries {
				var remove *DNSEntry
				switch entry.Type {
				case "A":
					if message.GetAddrFamily() == netlink.IPV4 {
						remove = entry
						remove.Content = message.GetIpNet().IP.String()
					}
				case "AAAA":
					if message.GetAddrFamily() == netlink.IPV6 {
						remove = entry
						remove.Content = message.GetIpNet().IP.String()
					}
				}
				if nil != remove {
					if _, err := t.client.DeleteDNSEntry(domain, remove); err != nil {
						return err
					} else {
						t.logger.Debug(fmt.Sprintf("removed %s.%s %s record %s", remove.Name, domain, remove.Type, entry.Content))
						remove.Content = ""
					}
				}
			}
		}
	}
	return nil
}
