package netlink

import (
	"golang.org/x/sys/unix"
	"net"
	"syscall"
	"unsafe"
)

func GetInterfaceLabelFromRouteAttr(attributes []syscall.NetlinkRouteAttr) string {
	for _, attribute := range attributes {
		if attribute.Attr.Type == unix.IFA_LABEL {
			return string(attribute.Value[:len(attribute.Value)-1])
		}
	}
	return ""
}

func GetCacheInfoFromRouteAttr(attributes []syscall.NetlinkRouteAttr) *unix.IfaCacheinfo {
	for _, attribute := range attributes {
		if attribute.Attr.Type == unix.IFA_CACHEINFO {
			return (*unix.IfaCacheinfo)(unsafe.Pointer(&attribute.Value[0:unix.SizeofIfaCacheinfo][0]))
		}
	}
	return nil
}

func GetIpNetFromRouteAttr(attributes *[]syscall.NetlinkRouteAttr, ifMsg *unix.IfAddrmsg) *net.IPNet {

	var dst *net.IPNet
	var loc *net.IPNet

	for _, attribute := range *attributes {
		switch attribute.Attr.Type {
		case unix.IFA_ADDRESS:
			dst = &net.IPNet{
				IP:   attribute.Value,
				Mask: net.CIDRMask(int(ifMsg.Prefixlen), 8*len(attribute.Value)),
			}
		case unix.IFA_LOCAL:
			var n = 8 * len(attribute.Value)
			loc = &net.IPNet{
				IP:   attribute.Value,
				Mask: net.CIDRMask(n, n),
			}
		}
	}

	if loc != nil && (ifMsg.Family != unix.AF_INET || false == loc.IP.Equal(dst.IP)) {
		return loc
	}

	return dst
}
