package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pbergman/nacl/netlink"
	"golang.org/x/sys/unix"
)

func main() {

	flag.Parse()

	var trap = make(chan os.Signal, 1)
	var ctx = context.Background()

	app, err := bootstrap(configFile)

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	link, err := netlink.ListenNetlink()

	if err != nil {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}

	defer link.Close()
	defer app.Close()

	signal.Notify(trap, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	app.Worker.Start(ctx, app.Config.Workers)

	var fdSet = new(unix.FdSet)
	var tval = unix.NsecToTimeval((100 * time.Millisecond).Nanoseconds())

mainApp:
	for {
		select {
		case x := <-trap:
			app.Logger.Debug(fmt.Sprintf("signal '%s' recieved, exiting", x.String()))
			app.Worker.Close()
			break mainApp
		default:
			ok, err := link.Wait(fdSet, &tval)

			if err != nil {

				if v, ok := err.(unix.Errno); ok && v.Temporary() {
					continue
				}

				app.Logger.Error(err)
			}

			if ok {
				messages, err := link.ReadMessages()

				if err != nil {
					fmt.Printf("failed reading netlink messages: %s", err)
					continue
				}

				_ = app.Worker.Queue(NewNetlinkMessagesJob(messages, app))
			}
		}
	}

	app.Logger.Debug("waiting for background processes to finish")
	app.Worker.Wait()
}
