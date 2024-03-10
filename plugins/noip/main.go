package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"github.com/pbergman/nacl/plugin"
	"golang.org/x/sys/unix"
)

var (
	Version string
	Name    plugin.PluginName    = "noip"
	Factory plugin.PluginFactory = func(config toml.Primitive, meta *toml.MetaData, logger *logger.Logger) (plugin.Plugin, error) {
		var conf *Config

		if err := meta.PrimitiveDecode(config, &conf); err != nil {
			return nil, err
		}

		return &NoIp{logger: logger}, nil
	}
)

type Group struct {
	Hostname string
	Username string
	Password string
	DDNSHost string `toml:"ddns_host"`
	Ipv6     bool
	offline  bool
}

type Config struct {
	Groups []*Group
}

type NoIp struct {
	logger *logger.Logger
	config *Config
}

func (n *NoIp) do(vars *url.Values, group *Group) error {
	request, err := http.NewRequest("GET", fmt.Sprintf("https://%s/nic/update?%s", group.DDNSHost, vars.Encode()), nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(group.Username, group.Password)
	request.Header.Set("user-agent", fmt.Sprintf("private nacl/1.0 pbergman@live.nl"))
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		n.logger.Error(fmt.Sprintf("%s %s %s", request.Method, request.URL.Path, request.Proto))
		return err
	}
	n.logger.Debug(fmt.Sprintf("%s %s %s %d", request.Method, request.URL.Path, request.Proto, resp.StatusCode))
	return nil
}

func (n *NoIp) Init(ip *net.IPNet) error {
	for _, group := range n.config.Groups {
		var vars = new(url.Values)
		if ipv4 := ip.IP.To4(); ipv4 != nil && len(ipv4) == net.IPv4len {
			vars.Set("hostname", group.Hostname)
			vars.Set("myip", ipv4.String())
			if err := n.do(vars, group); err != nil {
				n.logger.Error(err)
			}
			continue
		}
		if ipv6 := ip.IP.To16(); ipv6 != nil && len(ipv6) == net.IPv6len {
			if false == group.Ipv6 {
				continue
			}
			vars.Set("hostname", group.Hostname)
			vars.Set("myipv6", ipv6.String())
			if err := n.do(vars, group); err != nil {
				n.logger.Error(err)
			}
			continue
		}
	}
	return nil
}

func (n *NoIp) Handle(message *netlink.NetlinkMessageContext) error {
	switch message.GetMessageType() {
	case unix.RTM_NEWADDR:
		for _, group := range n.config.Groups {
			var vars = new(url.Values)
			vars.Set("hostname", group.Hostname)
			if message.GetAddrFamily() == netlink.IPV4 {
				vars.Set("myip", message.GetIpNet().IP.To4().String())

			}
			if message.GetAddrFamily() == netlink.IPV6 {
				if group.Ipv6 {
					vars.Set("myipv6", message.GetIpNet().IP.To16().String())
				} else {
					continue
				}
			}
			if group.offline {
				vars.Set("offline", "NO")
			}
			if err := n.do(vars, group); err != nil {
				n.logger.Error(err)
			}
		}
	case unix.RTM_DELADDR:
		for _, group := range n.config.Groups {
			if false == group.offline {
				continue
			}
			var vars = new(url.Values)
			vars.Set("hostname", group.Hostname)
			vars.Set("offline", "YES")
			if err := n.do(vars, group); err != nil {
				n.logger.Error(err)
			}
		}

	}
	return nil
}
