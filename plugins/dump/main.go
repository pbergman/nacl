package main

import (
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"github.com/pbergman/nacl/plugin"
	"golang.org/x/sys/unix"
)

var (
	Version string
	Name    plugin.PluginName    = "dump"
	Factory plugin.PluginFactory = func(config toml.Primitive, meta *toml.MetaData, logger *logger.Logger) (plugin.Plugin, error) {
		return &Dumper{logger: logger}, nil
	}
)

type Dumper struct {
	logger *logger.Logger
}

func (d *Dumper) Init(ip *net.IPNet) error {

	if ipv4 := ip.IP.To4(); ipv4 != nil && len(ipv4) == net.IPv4len {
		d.logger.Debug(fmt.Sprintf("[>>>] ipv4: %s", ipv4))
	} else {
		d.logger.Debug(fmt.Sprintf("[>>>] ipv6: %s", ip.IP.To16()))
	}

	return nil
}

func (d *Dumper) Handle(message *netlink.NetlinkMessageContext) error {

	var str string

	switch x := message.GetMessageType(); x {
	case unix.RTM_DELADDR:
		str = "[DEL] "
	case unix.RTM_NEWADDR:
		str = "[NEW] "
	default:
		str = fmt.Sprintf("[0x%02x] ", x)
	}

	switch message.GetAddrFamily() {
	case unix.AF_INET:
		str += "ipv4: %s"
	case unix.AF_INET6:
		str += "ipv6: %s"
	default:
		return nil
	}

	d.logger.Debug(
		logger.Message(
			fmt.Sprintf(str, message.GetIpNet().IP),
			map[string]interface{}{
				"pid":   message.GetMessagePid(),
				"scope": message.GetAddrScope(),
				"index": message.GetAddrIndex(),
				"seq":   message.GetMessageSequence(),
				"flags": message.GetMessageFlags(),
			},
		),
	)

	return nil
}
