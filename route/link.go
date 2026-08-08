// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package route

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// Link contains the configuration attributes for a network link.
//
// Not all fields will be populated by all link types, and may additionally be
// filtered based on the request message when a [LinkFilter] has been applied.
//
// References:
//   - linux/include/uapi/linux/if_link.h
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-attribute-set-link-attrs
type Link struct {
	Family           uint8
	Type             LinkType
	Index            int
	Flags            LinkFlags
	Address          net.HardwareAddr
	Broadcast        net.HardwareAddr
	Name             string
	MTU              uint32
	Link             uint32
	Qdisc            string
	Stats            *LinkStats
	Cost             string
	Priority         string
	Master           uint32
	Wireless         string
	ProtInfo         string
	TxQlen           uint32
	Map              *LinkIfMap
	Weight           uint32
	OperState        LinkOperState
	Mode             LinkMode
	Info             LinkInfo
	NetNsPid         uint32
	Alias            string
	NumVf            uint32
	Group            uint32
	NetNsFd          uint32
	ExtMask          uint32
	Promiscuity      uint32
	NumTxQueues      uint32
	NumRxQueues      uint32
	Carrier          uint8
	PhysPortId       int32
	CarrierChanges   uint32
	PhysSwitchId     int32
	LinkNetNsId      int32
	PhysPortName     string
	ProtoDown        uint8
	GsoMaxSegs       uint32
	GsoMaxSize       uint32
	XDP              *LinkXDPAttrs
	Event            uint32
	NewNetNsId       int32
	TargetNetNsId    int32
	CarrierUpCount   uint32
	CarrierDownCount uint32
	NewIfIndex       int32
	MinMTU           uint32
	MaxMTU           uint32
	AltNames         []string
	PermAddress      net.HardwareAddr
	ProtoDownReason  string
	ParentDevName    string
	ParentDevBusName string
	GroMaxSize       uint32
	TsoMaxSize       uint32
	TsoMaxSegs       uint32
	AllMulti         uint32
	DevLinkPort      string
	GsoIpv4MaxSize   uint32
	GroIpv4MaxSize   uint32
}

// UnmarshalNetlink unmarshals a Netlink message containing a [Link].
func (l *Link) UnmarshalNetlink(msg *netlink.Message) error {
	info, err := readIfInfomsg(msg)
	if err != nil {
		return fmt.Errorf("could not read IfInfomsg: %w", err)
	}

	l.Family = info.Family
	l.Type = LinkType(info.Type)
	l.Index = int(info.Index)
	l.Flags = LinkFlags(info.Flags)

	err = msg.Unmarshal(l)
	if err != nil {
		return fmt.Errorf("could not read attributed: %w", err)
	}

	return nil
}

// UnmarshalAttributes unmarshals the Netlink message attributes containing a
// [Link].
//
// The [IfInfomsg] header must already have been read.
func (l *Link) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_ADDRESS:
			l.Address = attr.Bytes()
		case unix.IFLA_BROADCAST:
			l.Broadcast = attr.Bytes()
		case unix.IFLA_IFNAME:
			l.Name = attr.String()
		case unix.IFLA_MTU:
			l.MTU = attr.Uint32()
		case unix.IFLA_LINK:
			l.Link = attr.Uint32()
		case unix.IFLA_QDISC:
			l.Qdisc = attr.String()
		case unix.IFLA_STATS, unix.IFLA_STATS64:
			l.Stats = new(LinkStats)
			err := attr.UnmarshalBytes(l.Stats)
			if err != nil {
				return err
			}
		case unix.IFLA_COST:
			l.Cost = attr.String()
		case unix.IFLA_PRIORITY:
			l.Priority = attr.String()
		case unix.IFLA_MASTER:
			l.Master = attr.Uint32()
		case unix.IFLA_WIRELESS:
			l.Wireless = attr.String()
		case unix.IFLA_PROTINFO:
			l.ProtInfo = attr.String()
		case unix.IFLA_TXQLEN:
			l.TxQlen = attr.Uint32()
		case unix.IFLA_MAP:
			l.Map = new(LinkIfMap)
			err := attr.UnmarshalBytes(l.Map)
			if err != nil {
				return err
			}
		case unix.IFLA_WEIGHT:
			l.Weight = attr.Uint32()
		case unix.IFLA_OPERSTATE:
			l.OperState = LinkOperState(attr.Uint8())
		case unix.IFLA_LINKMODE:
			l.Mode = LinkMode(attr.Uint8())
		case unix.IFLA_LINKINFO:
			err := attr.Unmarshal(&l.Info)
			if err != nil {
				return err
			}
		case unix.IFLA_NET_NS_PID:
			l.NetNsPid = attr.Uint32()
		case unix.IFLA_IFALIAS:
			l.Alias = attr.String()
		case unix.IFLA_NUM_VF:
			l.NumVf = attr.Uint32()
		case unix.IFLA_GROUP:
			l.Group = attr.Uint32()
		case unix.IFLA_NET_NS_FD:
			l.NetNsFd = attr.Uint32()
		case unix.IFLA_EXT_MASK:
			l.ExtMask = attr.Uint32()
		case unix.IFLA_PROMISCUITY:
			l.Promiscuity = attr.Uint32()
		case unix.IFLA_NUM_TX_QUEUES:
			l.NumTxQueues = attr.Uint32()
		case unix.IFLA_NUM_RX_QUEUES:
			l.NumRxQueues = attr.Uint32()
		case unix.IFLA_CARRIER:
			l.Carrier = attr.Uint8()
		case unix.IFLA_PHYS_PORT_ID:
			l.PhysPortId = attr.Int32()
		case unix.IFLA_CARRIER_CHANGES:
			l.CarrierChanges = attr.Uint32()
		case unix.IFLA_PHYS_SWITCH_ID:
			l.PhysSwitchId = attr.Int32()
		case unix.IFLA_LINK_NETNSID:
			l.LinkNetNsId = attr.Int32()
		case unix.IFLA_PHYS_PORT_NAME:
			l.PhysPortName = attr.String()
		case unix.IFLA_PROTO_DOWN:
			l.ProtoDown = attr.Uint8()
		case unix.IFLA_GSO_MAX_SEGS:
			l.GsoMaxSegs = attr.Uint32()
		case unix.IFLA_GSO_MAX_SIZE:
			l.GsoMaxSize = attr.Uint32()
		case unix.IFLA_XDP:
			l.XDP = new(LinkXDPAttrs)
			err := attr.Unmarshal(l.XDP)
			if err != nil {
				return err
			}
		case unix.IFLA_EVENT:
			l.Event = attr.Uint32()
		case unix.IFLA_NEW_NETNSID:
			l.NewNetNsId = attr.Int32()
		case unix.IFLA_TARGET_NETNSID:
			l.TargetNetNsId = attr.Int32()
		case unix.IFLA_CARRIER_UP_COUNT:
			l.CarrierUpCount = attr.Uint32()
		case unix.IFLA_CARRIER_DOWN_COUNT:
			l.CarrierDownCount = attr.Uint32()
		case unix.IFLA_NEW_IFINDEX:
			l.NewIfIndex = attr.Int32()
		case unix.IFLA_MIN_MTU:
			l.MinMTU = attr.Uint32()
		case unix.IFLA_MAX_MTU:
			l.MaxMTU = attr.Uint32()
		case unix.IFLA_ALT_IFNAME:
			l.AltNames = append(l.AltNames, attr.String())
		case unix.IFLA_PERM_ADDRESS:
			l.PermAddress = attr.Bytes()
		case unix.IFLA_PROTO_DOWN_REASON:
			l.ProtoDownReason = attr.String()
		case unix.IFLA_PARENT_DEV_NAME:
			l.ParentDevName = attr.String()
		case unix.IFLA_PARENT_DEV_BUS_NAME:
			l.ParentDevBusName = attr.String()
		case unix.IFLA_GRO_MAX_SIZE:
			l.GroMaxSize = attr.Uint32()
		case unix.IFLA_TSO_MAX_SIZE:
			l.TsoMaxSize = attr.Uint32()
		case unix.IFLA_TSO_MAX_SEGS:
			l.TsoMaxSegs = attr.Uint32()
		case unix.IFLA_ALLMULTI:
			l.AllMulti = attr.Uint32()
		case unix.IFLA_DEVLINK_PORT:
			l.DevLinkPort = attr.String()
		case unix.IFLA_GSO_IPV4_MAX_SIZE:
			l.GsoIpv4MaxSize = attr.Uint32()
		case unix.IFLA_GRO_IPV4_MAX_SIZE:
			l.GroIpv4MaxSize = attr.Uint32()
		}
	}

	// ensure link info is always set, even when a physical or intrinsic link
	// with no driver kind to reference.
	if l.Info.Kind == "" {
		l.Info.Kind = "device"
	}

	return nil
}

// Links contains a list of [Link] returns as a dump from a series of Netlink
// messages.
type Links []*Link

// UnmarshalNetlink unmarshals one-or-more Netlink messages containing a [Link]
// and appends it to itself.
func (ls *Links) UnmarshalNetlink(msg *netlink.Message) error {
	link := &Link{}

	err := link.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	*ls = append(*ls, link)
	return nil
}

// LinkDriver is contained within a [LinkInfo] for link types where additional
// information is made about the link by the driver.
//
// References:
//   - linux/drivers/net
type LinkDriver interface {
	// DriverName returns the name of the Linux driver that manages the link
	// device.
	DriverName() string
}

// LinkFilter is used for filtering the attributes within [LinkAttrs].
//
// References:
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-definition-rtext-filter
type LinkFilter uint32

// Constants for [LinkFilter].
const (
	LinkFilterVf LinkFilter = 1 << iota
	LinkFilterBrvlan
	LinkFilterBrvlanCompressed
	LinkFilterSkipStats
	LinkFilterMrp
	LinkFilterCfmConfig
	LinkFilterCfmStatus
	LinkFilterMst
)

// linkFilterNames is bit shifted through by [LinkFilter.String] to build a
// stringified representation.
var linkFilterNames = []string{
	"VF",
	"BRVLAN",
	"BRVLAN-COMPRESSED",
	"SKIP-STATS",
	"MRP",
	"CFM-CONFIG",
	"CFM-STATUS",
	"MST",
}

func (lf LinkFilter) String() string {
	var s strings.Builder

	for i, name := range linkFilterNames {
		if lf&(1<<i) != 0 {
			s.WriteByte(' ')
			s.WriteString(name)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}

// LinkFlags contains the status of a [Link].
//
// Note: these flags are not interchangeable with [net.Flags].
//
// References:
//   - linux/include/uapi/linux/if.h
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-definition-ifinfo-flags
type LinkFlags uint32

// Constants for [LinkFlags].
const (
	FlagUp LinkFlags = 1 << iota
	FlagBroadcast
	FlagDebug
	FlagLoopback
	FlagPointToPoint
	FlagNoTrailers
	FlagRunning
	FlagNoArp
	FlagPromisc
	FlagAllMulti
	FlagMaster
	FlagSlave
	FlagMulticast
	FlagPortsel
	FlagAutoMedia
	FlagDynamic
	FlagLowerUp
	FlagDormant
	FlagEcho
)

// linkFlagNames is bit shifted through by [LinkFlags.String] to build a
// stringified representation.
var linkFlagNames = []string{
	"UP",
	"BROADCAST",
	"DEBUG",
	"LOOPBACK",
	"POINT-TO-POINT",
	"NO-TRAILERS",
	"RUNNING",
	"NO-ARP",
	"PROMISC",
	"ALL-MULTI",
	"MASTER",
	"SLAVE",
	"MULTICAST",
	"PORTSEL",
	"AUTO-MEDIA",
	"DYNAMIC",
	"LOWER-UP",
	"DORMANT",
	"ECHO",
}

func (lf LinkFlags) String() string {
	var s strings.Builder

	for i, name := range linkFlagNames {
		if lf&(1<<i) != 0 {
			s.WriteByte(' ')
			s.WriteString(name)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}

// LinkIfMap contains information about memory and device configuration.
//
// References:
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-definition-rtnl-link-ifmap
type LinkIfMap struct {
	MemStart uint64
	MemEnd   uint64
	BaseAddr uint64
	IRQ      uint16
	DMA      uint8
	Port     uint8
}

// UnmarshalBinary unmarshals [LinkIfMap] from a [LinkAttrs] attribute.
//
// It will ignore any additional bytes it is given.
func (l *LinkIfMap) UnmarshalBinary(b []byte) error {
	if len(b) < 28 {
		return fmt.Errorf("rt-link-ifmap: expected at least 28 bytes, got %d", len(b))
	}

	l.MemStart = binary.NativeEndian.Uint64(b)
	l.MemEnd = binary.NativeEndian.Uint64(b[8:])
	l.BaseAddr = binary.NativeEndian.Uint64(b[16:])
	l.IRQ = binary.NativeEndian.Uint16(b[24:])
	l.DMA = b[26]
	l.Port = b[27]

	return nil
}

// LinkInfo contains the kind of link or link slave, as well as driver-specific
// information.
//
// References:
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-attribute-set-linkinfo-attrs
type LinkInfo struct {
	Kind      string
	Driver    LinkDriver
	SlaveKind string
}

// UnmarshalAttributes unmarshals [LinkInfo] from inside a [Link] attributes.
func (li *LinkInfo) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_INFO_KIND:
			li.Kind = attr.String()

			switch li.Kind {
			case "bareudp":
				li.Driver = &BareUDP{}

			case "bond":
				li.Driver = &Bond{}

			case "ipip":
				li.Driver = &IPIP{}
			}

		case unix.IFLA_INFO_DATA:
			// only unmarshal driver information if implemented.
			if au, ok := li.Driver.(netlink.AttributeUnmarshaler); ok {
				err := attr.Unmarshal(au)
				if err != nil {
					return fmt.Errorf("link-info-data: %w", err)
				}
			}

		case unix.IFLA_INFO_SLAVE_KIND:
			li.SlaveKind = attr.String()
		}
	}

	return nil
}

// LinkMode contains the mode of the link.
//
// References:
//   - linux/include/uapi/linux/if.h
type LinkMode uint8

// Constants for [LinkMode].
const (
	LinkModeDefault LinkMode = iota
	LinkModeDormant
	LinkModeTesting
)

func (lm LinkMode) String() string {
	switch lm {
	case LinkModeDefault:
		return "DEFAULT"
	case LinkModeDormant:
		return "DORMANT"
	case LinkModeTesting:
		return "TESTING"

	default:
		return "UNKNOWN"
	}
}

// LinkOperState contains the RFC 2864 operational status of a [Link].
//
// References:
//   - linux/include/uapi/linux/if.h
type LinkOperState uint8

// Constants for [LinkOperState].
const (
	LinkOperStateUnknown LinkOperState = iota
	LinkOperStateNotPresent
	LinkOperStateDown
	LinkOperStateLowerLayerDown
	LinkOperStateTesting
	LinkOperStateDormant
	LinkOperStateUp
)

func (los LinkOperState) String() string {
	switch los {
	case LinkOperStateNotPresent:
		return "NOT-PRESENT"
	case LinkOperStateDown:
		return "DOWN"
	case LinkOperStateLowerLayerDown:
		return "LOWER-LAYER-DOWN"
	case LinkOperStateTesting:
		return "TESTING"
	case LinkOperStateDormant:
		return "DORMANT"
	case LinkOperStateUp:
		return "UP"

	default:
		return "UNKNOWN"
	}
}

// LinkStats contains statistics for a link.
//
// This structure contains an amalgamation of both rtnl_link_stats and
// rtnl_link_stats64.
//
// References:
//   - linux/include/uapi/linux/if_link.h
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-definition-rtnl-link-stats
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-definition-rtnl-link-stats64
type LinkStats struct {
	RxPackets  uint64
	TxPackets  uint64
	RxBytes    uint64
	TxBytes    uint64
	RxErrors   uint64
	TxErrors   uint64
	RxDropped  uint64
	TxDropped  uint64
	Multicast  uint64
	Collisions uint64

	RxLengthErrors uint64
	RxOverErrors   uint64
	RxCrcErrors    uint64
	RxFrameErrors  uint64
	RxFifoErrors   uint64
	RxMissedErrors uint64

	TxAbortedErrors   uint64
	TxCarrierErrors   uint64
	TxFifoErrors      uint64
	TxHeartbeatErrors uint64
	TxWindowErrors    uint64

	RxCompressed uint64
	TxCompressed uint64

	// Linux 4.6+.
	RxNoHandler uint64

	// Linux 5.9+, 64-bit only.
	RxOtherHostDropped uint64
}

// UnmarshalBinary unmarshals either 32-bit or 32-bit link statistics.
//
// It will ignore any additional bytes it is given.
func (ls *LinkStats) UnmarshalBinary(b []byte) error {
	if len(b) < 92 {
		// cannot be either 32-bit or 64-bit statistics.
		return fmt.Errorf("rtnl-link-stats: expected at least 92 bytes, got %d", len(b))
	}

	if len(b) >= 184 {
		// minimum size of 64-bit statistics, greater than 32-bit statistics.
		return ls.UnmarshalBinary64(b)
	}

	return ls.UnmarshalBinary32(b)
}

// UnmarshalBinary32 unmarshals 32-bit link statistics.
//
// It will ignore any additional bytes it is given.
func (ls *LinkStats) UnmarshalBinary32(b []byte) error {
	if len(b) < 92 {
		return fmt.Errorf("rtnl-link-stats: expected at least 92 bytes, got %d", len(b))
	}

	ls.RxPackets = uint64(binary.NativeEndian.Uint32(b))
	ls.TxPackets = uint64(binary.NativeEndian.Uint32(b[4:]))
	ls.RxBytes = uint64(binary.NativeEndian.Uint32(b[8:]))
	ls.TxBytes = uint64(binary.NativeEndian.Uint32(b[12:]))
	ls.RxErrors = uint64(binary.NativeEndian.Uint32(b[16:]))
	ls.TxErrors = uint64(binary.NativeEndian.Uint32(b[20:]))
	ls.RxDropped = uint64(binary.NativeEndian.Uint32(b[24:]))
	ls.TxDropped = uint64(binary.NativeEndian.Uint32(b[28:]))
	ls.Multicast = uint64(binary.NativeEndian.Uint32(b[32:]))
	ls.Collisions = uint64(binary.NativeEndian.Uint32(b[36:]))

	ls.RxLengthErrors = uint64(binary.NativeEndian.Uint32(b[40:]))
	ls.RxOverErrors = uint64(binary.NativeEndian.Uint32(b[44:]))
	ls.RxCrcErrors = uint64(binary.NativeEndian.Uint32(b[48:]))
	ls.RxFrameErrors = uint64(binary.NativeEndian.Uint32(b[52:]))
	ls.RxFifoErrors = uint64(binary.NativeEndian.Uint32(b[56:]))
	ls.RxMissedErrors = uint64(binary.NativeEndian.Uint32(b[60:]))

	ls.TxAbortedErrors = uint64(binary.NativeEndian.Uint32(b[64:]))
	ls.TxCarrierErrors = uint64(binary.NativeEndian.Uint32(b[68:]))
	ls.TxFifoErrors = uint64(binary.NativeEndian.Uint32(b[72:]))
	ls.TxHeartbeatErrors = uint64(binary.NativeEndian.Uint32(b[76:]))
	ls.TxWindowErrors = uint64(binary.NativeEndian.Uint32(b[80:]))

	ls.RxCompressed = uint64(binary.NativeEndian.Uint32(b[84:]))
	ls.TxCompressed = uint64(binary.NativeEndian.Uint32(b[88:]))

	if len(b) >= 96 {
		ls.RxNoHandler = uint64(binary.NativeEndian.Uint32(b[92:]))
	}

	return nil
}

// UnmarshalBinary64 unmarshals 64-bit link statistics.
//
// It will ignore any additional bytes it is given.
func (ls *LinkStats) UnmarshalBinary64(b []byte) error {
	if len(b) < 184 {
		return fmt.Errorf("rtnl-link-stats-64: expected at least 184 bytes, got %d", len(b))
	}

	ls.RxPackets = binary.NativeEndian.Uint64(b)
	ls.TxPackets = binary.NativeEndian.Uint64(b[8:])
	ls.RxBytes = binary.NativeEndian.Uint64(b[16:])
	ls.TxBytes = binary.NativeEndian.Uint64(b[24:])
	ls.RxErrors = binary.NativeEndian.Uint64(b[32:])
	ls.TxErrors = binary.NativeEndian.Uint64(b[40:])
	ls.RxDropped = binary.NativeEndian.Uint64(b[48:])
	ls.TxDropped = binary.NativeEndian.Uint64(b[56:])
	ls.Multicast = binary.NativeEndian.Uint64(b[64:])
	ls.Collisions = binary.NativeEndian.Uint64(b[72:])

	ls.RxLengthErrors = binary.NativeEndian.Uint64(b[80:])
	ls.RxOverErrors = binary.NativeEndian.Uint64(b[88:])
	ls.RxCrcErrors = binary.NativeEndian.Uint64(b[96:])
	ls.RxFrameErrors = binary.NativeEndian.Uint64(b[104:])
	ls.RxFifoErrors = binary.NativeEndian.Uint64(b[112:])
	ls.RxMissedErrors = binary.NativeEndian.Uint64(b[120:])

	ls.TxAbortedErrors = binary.NativeEndian.Uint64(b[128:])
	ls.TxCarrierErrors = binary.NativeEndian.Uint64(b[136:])
	ls.TxFifoErrors = binary.NativeEndian.Uint64(b[144:])
	ls.TxHeartbeatErrors = binary.NativeEndian.Uint64(b[152:])
	ls.TxWindowErrors = binary.NativeEndian.Uint64(b[160:])

	ls.RxCompressed = binary.NativeEndian.Uint64(b[168:])
	ls.TxCompressed = binary.NativeEndian.Uint64(b[176:])

	if len(b) >= 192 {
		ls.RxNoHandler = binary.NativeEndian.Uint64(b[184:])
	}

	if len(b) >= 200 {
		ls.RxOtherHostDropped = binary.NativeEndian.Uint64(b[192:])
	}

	return nil
}

// LinkType is the wire type of a [Link], defined by it's ARP hardware type.
type LinkType uint16

func (lt LinkType) String() string {
	switch lt {
	case unix.ARPHRD_6LOWPAN:
		return "6lowpan"
	case unix.ARPHRD_ADAPT:
		return "adapt"
	case unix.ARPHRD_APPLETLK:
		return "appletlk"
	case unix.ARPHRD_ARCNET:
		return "arcnet"
	case unix.ARPHRD_ASH:
		return "ash"
	case unix.ARPHRD_ATM:
		return "atm"
	case unix.ARPHRD_AX25:
		return "ax25"
	case unix.ARPHRD_BIF:
		return "bif"
	case unix.ARPHRD_CAIF:
		return "caif"
	case unix.ARPHRD_CAN:
		return "can"
	case unix.ARPHRD_CHAOS:
		return "chaos"
	case unix.ARPHRD_CSLIP:
		return "cslip"
	case unix.ARPHRD_CSLIP6:
		return "cslip6"
	case unix.ARPHRD_DDCMP:
		return "ddcmp"
	case unix.ARPHRD_DLCI:
		return "dlci"
	case unix.ARPHRD_ECONET:
		return "econet"
	case unix.ARPHRD_EETHER:
		return "eether"
	case unix.ARPHRD_ETHER:
		return "ether"
	case unix.ARPHRD_EUI64:
		return "eui64"
	case unix.ARPHRD_FCAL:
		return "fcal"
	case unix.ARPHRD_FCFABRIC:
		return "fcfabric"
	case unix.ARPHRD_FCPL:
		return "fcpl"
	case unix.ARPHRD_FCPP:
		return "fcpp"
	case unix.ARPHRD_FDDI:
		return "fddi"
	case unix.ARPHRD_FRAD:
		return "frad"
	case unix.ARPHRD_HIPPI:
		return "hippi"
	case unix.ARPHRD_HWX25:
		return "hwx25"
	case unix.ARPHRD_IEEE1394:
		return "ieee1394"
	case unix.ARPHRD_IEEE802:
		return "ieee802"
	case unix.ARPHRD_IEEE80211:
		return "ieee80211"
	case unix.ARPHRD_IEEE80211_PRISM:
		return "ieee80211_prism"
	case unix.ARPHRD_IEEE80211_RADIOTAP:
		return "ieee80211_radiotap"
	case unix.ARPHRD_IEEE802154:
		return "ieee802154"
	case unix.ARPHRD_IEEE802154_MONITOR:
		return "ieee802154_monitor"
	case unix.ARPHRD_IEEE802_TR:
		return "ieee802_tr"
	case unix.ARPHRD_INFINIBAND:
		return "infiniband"
	case unix.ARPHRD_IP6GRE:
		return "ip6gre"
	case unix.ARPHRD_IPDDP:
		return "ipddp"
	case unix.ARPHRD_IPGRE:
		return "ipgre"
	case unix.ARPHRD_IRDA:
		return "irda"
	case unix.ARPHRD_LAPB:
		return "lapb"
	case unix.ARPHRD_LOCALTLK:
		return "localtlk"
	case unix.ARPHRD_LOOPBACK:
		return "loopback"
	case unix.ARPHRD_MCTP:
		return "mctp"
	case unix.ARPHRD_METRICOM:
		return "metricom"
	case unix.ARPHRD_NETLINK:
		return "netlink"
	case unix.ARPHRD_NETROM:
		return "netrom"
	case unix.ARPHRD_NONE:
		return "none"
	case unix.ARPHRD_PHONET:
		return "phonet"
	case unix.ARPHRD_PHONET_PIPE:
		return "phonet_pipe"
	case unix.ARPHRD_PIMREG:
		return "pimreg"
	case unix.ARPHRD_PPP:
		return "ppp"
	case unix.ARPHRD_PRONET:
		return "pronet"
	case unix.ARPHRD_RAWHDLC:
		return "rawhdlc"
	case unix.ARPHRD_RAWIP:
		return "rawip"
	case unix.ARPHRD_ROSE:
		return "rose"
	case unix.ARPHRD_RSRVD:
		return "rsrvd"
	case unix.ARPHRD_SIT:
		return "sit"
	case unix.ARPHRD_SKIP:
		return "skip"
	case unix.ARPHRD_SLIP:
		return "slip"
	case unix.ARPHRD_SLIP6:
		return "slip6"
	case unix.ARPHRD_TUNNEL:
		return "tunnel"
	case unix.ARPHRD_TUNNEL6:
		return "tunnel6"
	case unix.ARPHRD_VOID:
		return "void"
	case unix.ARPHRD_VSOCKMON:
		return "vsockmon"
	case unix.ARPHRD_X25:
		return "x25"

	default:
		return "unknown"
	}
}

// LinkXDPAttrs contains information about an eXpress Data Path (XDP) program
// attached to a link.
//
// References:
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html#rt-link-attribute-set-xdp-attrs
type LinkXDPAttrs struct {
	Fd         int32
	Attached   uint8
	Flags      uint32
	ProgId     uint32
	DrvProgId  uint32
	SkbProgId  uint32
	HwProgId   uint32
	ExpectedFd uint32
}

// UnmarshalAttributes unmarshal an XDP link attributes.
func (x *LinkXDPAttrs) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_XDP_FD:
			x.Fd = attr.Int32()
		case unix.IFLA_XDP_ATTACHED:
			x.Attached = attr.Uint8()
		case unix.IFLA_XDP_FLAGS:
			x.Flags = attr.Uint32()
		case unix.IFLA_XDP_PROG_ID:
			x.ProgId = attr.Uint32()
		case unix.IFLA_XDP_DRV_PROG_ID:
			x.DrvProgId = attr.Uint32()
		case unix.IFLA_XDP_SKB_PROG_ID:
			x.SkbProgId = attr.Uint32()
		case unix.IFLA_XDP_HW_PROG_ID:
			x.HwProgId = attr.Uint32()
		case unix.IFLA_XDP_EXPECTED_FD:
			x.ExpectedFd = attr.Uint32()
		}
	}

	return nil
}
