package main

import (
	"context"

	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"github.com/pbergman/nacl/plugin"
)

type PluginJob struct {
	plugin  plugin.Plugin
	message *netlink.NetlinkMessageContext
}

func (p *PluginJob) Run(ctx context.Context, logger *logger.Logger) {
	select {
	case <-ctx.Done():

		if err := ctx.Err(); err != nil {
			logger.Error(err)
		}

		return
	default:

		if err := p.plugin.Handle(p.message); err != nil {
			logger.Error(err)
		}
	}
}
