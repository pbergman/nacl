# Netlink Address Change Listener

This application will listen for netlink notification when ip address on configured interfaces changes and will dispatch that to the enabled plugins so that the can for example update dns records.

I have created this application to run on an EdgeRouter because the dns provider does not support ddns. 

This application uses plugins for handling messages, so it should be easily extendable. See the [dump](plugins%2Fdump%2Fmain.go) plugin for reference.

### Config

The config file should be in a [toml](https://toml.io/en/) format and uses the [BurntSushi/toml](https://github.com/BurntSushi/toml) library for processing the config. 
 
```toml
## Required and should be the interface we 
## want to listen on 
interface="eth0"

## Enable to print debug messages, if false
## it will print only messages when an error
## accoured (with the last debug messages)
# debug=false

## By default it will read the /etc/ip-listener/plugins-enabled 
## dir for plugin but can be changed here 
# plugin_dir="..."

## This application wil create some threats (go routines) 
## for handling work like intializing plugins and 
## processing/dispatching netlink messages. By default
## this will be set the ammount of proccessors of
## this machine but can be set to other number.
# workers=4

## Plugins specific config which should a table prefixed 
## with plugin following the plugin name.
# [plugin.<NAME>]
# ...
```

### Plugins

The following plugins are provided with this application.
  
  - [dump](plugins%2Fdump%2FREADME.md)
  - [no ip](plugins%2Fnoip%2FREADME.md)
  - [transip](plugins%2Ftransip%2FREADME.md)

### Plugin

for creating new plugin you to create a global var Name (plugin.PluginName), Factory (plugin.PluginFactory) and optional Version (string) so the application can load them on boot. 

A basic example could be:

```go
package main

import (
	"net"
	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"github.com/pbergman/nacl/plugin"

)

var (
	Version plugin.PluginVersion = "dev"
	Name    plugin.PluginName    = "test"
	Factory plugin.PluginFactory = func(config toml.Primitive, meta *toml.MetaData, logger *logger.Logger) (plugin.Plugin, error) {
		return &Test{logger: logger}, nil
	}
)

type Test struct {
	logger *logger.Logger
}

func (t *Test) Init(ip *net.IPNet) error {
	t.logger.Debug("initializing plugin...")
	return nil
}

func (t *Test) Handle(message *netlink.NetlinkMessageContext) error {
	t.logger.Debug("update received")
	return nil
}
```

It is also possible to use config for your plugin as all defined config in the `[plugin.<PLUGIN_NAME>]` block would be received as a raw (toml.Primitive) which can serialized in your factory function to any desired data type (see [noip plugin](plugins%2Fnoip%2Fmain.go) for simple example).
