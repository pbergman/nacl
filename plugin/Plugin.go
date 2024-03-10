package plugin

import (
	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"net"
)

type (
	PluginName    string
	PluginFactory func(config toml.Primitive, meta *toml.MetaData, logger *logger.Logger) (Plugin, error)
	Plugin        interface {
		Handle(message *netlink.NetlinkMessageContext) error
		Init(ip *net.IPNet) error
	}
)
