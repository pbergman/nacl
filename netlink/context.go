package netlink

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IPV4 = unix.AF_INET
	IPV6 = unix.AF_INET6
)

func NewNetlinkContext(message *syscall.NetlinkMessage) *NetlinkMessageContext {
	return &NetlinkMessageContext{linkMessage: message}
}

type NetlinkMessageContext struct {
	linkMessage *syscall.NetlinkMessage
	addrMessage *unix.IfAddrmsg
	routeAttr   []syscall.NetlinkRouteAttr
}

func (n *NetlinkMessageContext) getAddrMessage() *unix.IfAddrmsg {

	if n.addrMessage == nil {
		n.addrMessage = (*unix.IfAddrmsg)(unsafe.Pointer(&n.linkMessage.Data[0:unix.SizeofIfAddrmsg][0]))
	}

	return n.addrMessage
}

func (n *NetlinkMessageContext) getRouteAttr() []syscall.NetlinkRouteAttr {

	if nil == n.routeAttr {
		attr, err := syscall.ParseNetlinkRouteAttr(n.linkMessage)

		if err != nil {
			// skip errors for now and return
			// successfully parsed values back
			return attr
		}

		n.routeAttr = attr
	}

	return n.routeAttr
}

func (n *NetlinkMessageContext) GetAddrFamily() uint8 {
	return n.getAddrMessage().Family
}

func (n *NetlinkMessageContext) GetAddrPrefixLen() uint8 {
	return n.getAddrMessage().Prefixlen
}

func (n *NetlinkMessageContext) GetAddrFlags() uint8 {
	return n.getAddrMessage().Flags
}

func (n *NetlinkMessageContext) GetAddrScope() uint8 {
	return n.getAddrMessage().Scope
}

func (n *NetlinkMessageContext) GetAddrIndex() uint32 {
	return n.getAddrMessage().Index
}

func (n *NetlinkMessageContext) IsNewAddr() bool {
	return n.GetMessageType() == unix.RTM_NEWADDR
}

func (n *NetlinkMessageContext) IsDelAddr() bool {
	return n.GetMessageType() == unix.RTM_DELADDR
}

func (n *NetlinkMessageContext) GetMessageType() uint16 {
	return n.linkMessage.Header.Type
}

func (n *NetlinkMessageContext) GetMessageError() error {
	var buf = [2]byte{}
	var err uint32

	if *(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD); buf == [2]byte{0xCD, 0xAB} {
		err = binary.LittleEndian.Uint32(n.linkMessage.Data[0:4])
	} else if buf == [2]byte{0xAB, 0xCD} {
		err = binary.BigEndian.Uint32(n.linkMessage.Data[0:4])
	}

	return unix.Errno(-err)
}

func (n *NetlinkMessageContext) GetMessageFlags() uint16 {
	return n.linkMessage.Header.Flags
}

func (n *NetlinkMessageContext) GetMessageSequence() uint32 {
	return n.linkMessage.Header.Seq
}

func (n *NetlinkMessageContext) GetMessagePid() uint32 {
	return n.linkMessage.Header.Pid
}

func (n *NetlinkMessageContext) GetInterfaceLabel() string {
	for _, attribute := range n.getRouteAttr() {
		if attribute.Attr.Type == unix.IFA_LABEL {
			return string(attribute.Value[:len(attribute.Value)-1])
		}
	}
	return ""
}

func (n *NetlinkMessageContext) GetCacheInfo() *unix.IfaCacheinfo {
	for _, attribute := range n.getRouteAttr() {
		if attribute.Attr.Type == unix.IFA_CACHEINFO {
			return (*unix.IfaCacheinfo)(unsafe.Pointer(&attribute.Value[0:unix.SizeofIfaCacheinfo][0]))
		}
	}
	return nil
}

func (n *NetlinkMessageContext) GetIpNet() *net.IPNet {

	var dst *net.IPNet
	var loc *net.IPNet

	for _, attribute := range n.getRouteAttr() {
		switch attribute.Attr.Type {
		case unix.IFA_ADDRESS:
			dst = &net.IPNet{
				IP:   attribute.Value,
				Mask: net.CIDRMask(int(n.addrMessage.Prefixlen), 8*len(attribute.Value)),
			}
		case unix.IFA_LOCAL:
			var n = 8 * len(attribute.Value)
			loc = &net.IPNet{
				IP:   attribute.Value,
				Mask: net.CIDRMask(n, n),
			}
		}
	}

	if loc != nil && (n.GetAddrFamily() != IPV4 || false == loc.IP.Equal(dst.IP)) {
		return loc
	}

	return dst
}
