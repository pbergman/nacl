package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pbergman/logger"
	linked "github.com/pbergman/nacl/plugin"
	"net"
	"path/filepath"
	"plugin"
	"runtime"
)

type Config struct {
	Debug     bool
	Workers   uint8
	Interface string
	PluginDir string `toml:"plugin_dir"`
}

func bootstrap(file string) (*App, error) {

	var conf struct {
		Config
		Plugins map[string]toml.Primitive `toml:"plugin"`
	}

	meta, err := toml.DecodeFile(file, &conf)

	if err != nil {
		return nil, err
	}

	if "" == conf.Config.PluginDir {
		conf.Config.PluginDir = "/etc/ip-listener/plugins-enabled"
	}

	if 0 == conf.Config.Workers {
		conf.Config.Workers = uint8(runtime.NumCPU())
	}

	var logger = GetLogger(&conf.Config)

	plugins, err := readPlugins(&conf.Config, meta, conf.Plugins, logger)

	if err != nil {
		return nil, err
	}

	link, err := net.InterfaceByName(conf.Interface)

	if err != nil {
		return nil, fmt.Errorf("net: %s", err)
	}

	addrs, err := link.Addrs()

	if err != nil {
		return nil, fmt.Errorf("net.address: %s", err)
	}

	for _, x := range addrs {

		if addr, ok := x.(*net.IPNet); ok {

			for _, x := range plugins {
				if err := x.Init(addr); err != nil {
					return nil, fmt.Errorf("plugin.init: %s", err)
				}
			}
		}
	}

	return &App{Config: &conf.Config, Plugins: plugins, Logger: logger, LinkId: link.Index, Worker: NewWorker(logger)}, nil
}

func readPlugins(config *Config, meta toml.MetaData, pConfig map[string]toml.Primitive, log *logger.Logger) (map[string]linked.Plugin, error) {
	var plugins = make(map[string]linked.Plugin)

	log.Debug(fmt.Sprintf("loading plugins from '%s'", filepath.Join(config.PluginDir, "*.so")))

	files, err := filepath.Glob(filepath.Join(config.PluginDir, "*.so"))

	if err != nil {
		return nil, err
	}

	for _, file := range files {

		so, err := plugin.Open(file)

		if err != nil {
			return nil, err
		}

		var name string

		if sym, err := lookupSymbol[*linked.PluginName]("Name", so); err != nil {
			return nil, err
		} else {
			name = string(*sym)
		}

		if sym, err := lookupSymbol[*linked.PluginFactory]("Factory", so); err != nil {
			return nil, err
		} else {
			var c toml.Primitive

			if v, ok := pConfig[name]; ok {
				c = v
			}

			p, err := (*sym)(c, &meta, log.WithName(name))

			if err != nil {
				return nil, err
			}

			plugins[name] = p

			var ctx = map[string]interface{}{
				"file": filepath.Base(file),
			}

			if sym, err := lookupSymbol[*linked.PluginName]("Version", so); err == nil {
				ctx["version"] = string(*sym)
			}

			log.Debug(logger.Message(fmt.Sprintf("loaded plugin '%s'", name), ctx))
		}
	}

	return plugins, nil
}

func lookupSymbol[T *linked.PluginName | *linked.PluginFactory | *linked.PluginVersion](n string, p *plugin.Plugin) (T, error) {
	symbol, err := p.Lookup(n)

	if err != nil {
		return nil, err
	}

	return symbol.(T), nil
}
