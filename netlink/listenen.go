package netlink

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"sync"
	"syscall"
)

func ListenNetlink() (*NetlinkListener, error) {

	socket, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_DGRAM, unix.NETLINK_ROUTE)

	if err != nil {
		return nil, fmt.Errorf("socket: %s", err)
	}

	var addr = &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Pid:    uint32(0),
		Groups: uint32((1 << (unix.RTNLGRP_LINK - 1)) | (1 << (unix.RTNLGRP_IPV4_IFADDR - 1)) | (1 << (unix.RTNLGRP_IPV6_IFADDR - 1))),
	}

	if err := unix.Bind(socket, addr); err != nil {
		return nil, fmt.Errorf("bind: %s", err)
	}

	var pool = &sync.Pool{
		New: func() any {
			return make([]byte, os.Getpagesize())
		},
	}

	return &NetlinkListener{fd: socket, addr: addr, pool: pool}, nil
}

type NetlinkListener struct {
	fd   int
	addr *unix.SockaddrNetlink
	pool *sync.Pool
}

func (l *NetlinkListener) Close() error {
	return unix.Close(l.fd)
}

func (l *NetlinkListener) Wait(fdSet *unix.FdSet, timeout *unix.Timeval) (bool, error) {

	if fdSet == nil {
		fdSet = new(unix.FdSet)
	}

	fdSet.Set(l.fd)

	n, err := unix.Select(l.fd+1, fdSet, nil, nil, timeout)

	if err != nil {
		return false, err
	}

	if n > 0 && fdSet.IsSet(l.fd) {
		return true, nil
	}

	return false, nil

}

func (l *NetlinkListener) ReadMessages() ([]syscall.NetlinkMessage, error) {
	defer func() {
		recover()
	}()

	var buf = l.pool.Get().([]byte)

	defer l.pool.Put(buf)

	n, err := unix.Read(l.fd, buf)

	if err != nil {
		return nil, fmt.Errorf("read: %s", err)
	}

	if n < unix.NLMSG_HDRLEN {
		return nil, fmt.Errorf("short response from netlink (%d)", n)
	}

	msgs, err := syscall.ParseNetlinkMessage(buf[:n])

	if err != nil {
		return nil, fmt.Errorf("parse: %s", err)
	}

	return msgs, nil
}
