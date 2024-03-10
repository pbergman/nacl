package main

import (
	"context"
	"fmt"
	"syscall"

	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/netlink"
	"golang.org/x/sys/unix"
)

func NewNetlinkMessagesJob(messages []syscall.NetlinkMessage, app *App) *NetlinkMessagesJob {

	var size = len(messages)
	var pool = make(chan *netlink.NetlinkMessageContext, len(messages))

	for i := 0; i < size; i++ {
		pool <- netlink.NewNetlinkContext(&messages[i])
	}

	close(pool)

	return &NetlinkMessagesJob{messages: pool, app: app}
}

type NetlinkMessagesJob struct {
	messages chan *netlink.NetlinkMessageContext
	app      *App
}

func (w *NetlinkMessagesJob) Run(ctx context.Context, logger *logger.Logger) {

	select {
	case <-ctx.Done():

		if err := ctx.Err(); err != nil {
			logger.Error(err)
		}

		return
	case message, ok := <-w.messages:

		if !ok {
			return
		}

		switch message.GetMessageType() {
		case unix.NLMSG_ERROR:
			logger.Error(fmt.Sprintf("error message: %v", message.GetMessageError()))
		case unix.RTM_DELADDR, unix.RTM_NEWADDR:

			if message.GetAddrIndex() != uint32(w.app.LinkId) {
				logger.Notice(fmt.Sprintf("valid message type recieved but wrong interface %d", message.GetAddrIndex()))
				break
			}

			for _, plugin := range w.app.Plugins {
				_ = w.app.Worker.Queue(&PluginJob{
					plugin:  plugin,
					message: message,
				})
			}

		case unix.NLMSG_DONE:
			break

		}
	}
}
