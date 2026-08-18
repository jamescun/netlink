// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// Family configures the family of internet protocol a [Link] has available.
type Family uint8

// Constants for [Family].
const (
	ALL   Family = unix.AF_UNSPEC
	INET  Family = unix.AF_INET
	INET6 Family = unix.AF_INET6
)

func (f Family) String() string {
	switch f {
	case ALL:
		return "ALL"
	case INET:
		return "INET"
	case INET6:
		return "INET6"

	default:
		return "UNKNOWN"
	}
}

// Flags contains the status of a [Link].
//
// Note: these flags are not interchangeable with [net.Flags].
//
// References:
//   - linux/include/uapi/linux/if.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-link.html#ifinfo-flags
type Flags uint32

// Constants for [Flags].
const (
	UP Flags = 1 << iota
	BROADCAST
	DEBUG
	LOOPBACK
	POINT_TO_POINT
	NO_TRAILERS
	RUNNING
	NO_ARP
	PROMISC
	ALL_MULTI
	MASTER
	SLAVE
	MULTICAST
	PORTSEL
	AUTO_MEDIA
	DYNAMIC
	LOWER_UP
	DORMANT
	ECHO
)

// flagNames is bit shifted through by [Flags.String] to build a stringified
// representation.
var flagNames = []string{
	"UP",
	"BROADCAST",
	"DEBUG",
	"LOOPBACK",
	"POINT_TO_POINT",
	"NO_TRAILERS",
	"RUNNING",
	"NO_ARP",
	"PROMISC",
	"ALL_MULTI",
	"MASTER",
	"SLAVE",
	"MULTICAST",
	"PORTSEL",
	"AUTO_MEDIA",
	"DYNAMIC",
	"LOWER_UP",
	"DORMANT",
	"ECHO",
}

func (f Flags) String() string {
	if f == 0 {
		return "NONE"
	}

	var s strings.Builder

	for i, name := range flagNames {
		if f&(1<<i) != 0 {
			s.WriteByte(' ')
			s.WriteString(name)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}

// IfInfoMsg is the fixed-length header before [Link] messages.
type IfInfoMsg struct {
	Family  Family
	Type    EtherType
	Index   int
	Flags   Flags
	Changed Flags
}

// Len returns the fixed-length of the [IfInfoMsg] header.
func (IfInfoMsg) Len() int { return 16 }

// AppendBinary appends an [IfInfoMsg] to bytes using the host byteorder.
func (i IfInfoMsg) AppendBinary(b []byte) ([]byte, error) {
	if i.Index < math.MinInt32 || i.Index > math.MaxInt32 {
		return nil, fmt.Errorf("index exceeds int32: %d", i.Index)
	}

	b = append(b, uint8(i.Family), 0x00)
	b = binary.NativeEndian.AppendUint16(b, uint16(i.Type))
	b = binary.NativeEndian.AppendUint32(b, uint32(i.Index)) //nolint
	b = binary.NativeEndian.AppendUint32(b, uint32(i.Flags))
	b = binary.NativeEndian.AppendUint32(b, uint32(i.Changed))

	return b, nil
}

// MarshalBinary marshals an [IfInfoMsg] to bytes using the host byteorder.
func (i IfInfoMsg) MarshalBinary() ([]byte, error) {
	return i.AppendBinary(make([]byte, 0, 16))
}

// MarshalNetlink marshals an [IfInfoMsg] to a Netlink message, for messages to
// rtnetlink that require only the header.
func (i IfInfoMsg) MarshalNetlink(msg netlink.MessageEncoder) error {
	err := msg.MarshalBytes(i)
	if err != nil {
		return fmt.Errorf("ifinfomsg: %w", err)
	}

	return nil
}

// UnmarshalBinary unmarshals an [IfInfoMsg] from bytes using the host
// byteorder.
//
// It will ignore any additional bytes it is given.
func (i *IfInfoMsg) UnmarshalBinary(b []byte) error {
	if len(b) < 16 {
		return fmt.Errorf("expected 16 bytes, got %d", len(b))
	}

	i.Family = Family(b[0])
	i.Type = EtherType(binary.NativeEndian.Uint16(b[2:]))
	i.Index = int(binary.NativeEndian.Uint32(b[4:]))
	i.Flags = Flags(binary.NativeEndian.Uint32(b[8:]))
	i.Changed = Flags(binary.NativeEndian.Uint32(b[12:]))

	return nil
}

// String returns a string representation of the attribute header for
// debugging.
func (i IfInfoMsg) String() string {
	return fmt.Sprintf(
		";; ->>IFINFOMSG<<- family: %d, type: %s, index: %d\n;; flags: %s, changed: %s\n",
		i.Family, i.Type, i.Index, i.Flags, i.Changed,
	)
}

// EtherType is the on-the-wire encapsulation type for a [Link] and tunneling
// devices, linked to the ARP Hardware Type.
type EtherType uint16

func (t EtherType) String() string {
	switch t {
	case unix.ARPHRD_6LOWPAN:
		return "6LOWPAN"
	case unix.ARPHRD_ADAPT:
		return "ADAPT"
	case unix.ARPHRD_APPLETLK:
		return "APPLETLK"
	case unix.ARPHRD_ARCNET:
		return "ARCNET"
	case unix.ARPHRD_ASH:
		return "ASH"
	case unix.ARPHRD_ATM:
		return "ATM"
	case unix.ARPHRD_AX25:
		return "AX25"
	case unix.ARPHRD_BIF:
		return "BIF"
	case unix.ARPHRD_CAIF:
		return "CAIF"
	case unix.ARPHRD_CAN:
		return "CAN"
	case unix.ARPHRD_CHAOS:
		return "CHAOS"
	case unix.ARPHRD_CSLIP:
		return "CSLIP"
	case unix.ARPHRD_CSLIP6:
		return "CSLIP6"
	case unix.ARPHRD_DDCMP:
		return "DDCMP"
	case unix.ARPHRD_DLCI:
		return "DLCI"
	case unix.ARPHRD_ECONET:
		return "ECONET"
	case unix.ARPHRD_EETHER:
		return "EETHER"
	case unix.ARPHRD_ETHER:
		return "ETHER"
	case unix.ARPHRD_EUI64:
		return "EUI64"
	case unix.ARPHRD_FCAL:
		return "FCAL"
	case unix.ARPHRD_FCFABRIC:
		return "FCFABRIC"
	case unix.ARPHRD_FCPL:
		return "FCPL"
	case unix.ARPHRD_FCPP:
		return "FCPP"
	case unix.ARPHRD_FDDI:
		return "FDDI"
	case unix.ARPHRD_FRAD:
		return "FRAD"
	case unix.ARPHRD_HIPPI:
		return "HIPPI"
	case unix.ARPHRD_HWX25:
		return "HWX25"
	case unix.ARPHRD_IEEE1394:
		return "IEEE1394"
	case unix.ARPHRD_IEEE802:
		return "IEEE802"
	case unix.ARPHRD_IEEE80211:
		return "IEEE80211"
	case unix.ARPHRD_IEEE80211_PRISM:
		return "IEEE80211_PRISM"
	case unix.ARPHRD_IEEE80211_RADIOTAP:
		return "IEEE80211_RADIOTAP"
	case unix.ARPHRD_IEEE802154:
		return "IEEE802154"
	case unix.ARPHRD_IEEE802154_MONITOR:
		return "IEEE802154_MONITOR"
	case unix.ARPHRD_IEEE802_TR:
		return "IEEE802_TR"
	case unix.ARPHRD_INFINIBAND:
		return "INFINIBAND"
	case unix.ARPHRD_IP6GRE:
		return "IP6GRE"
	case unix.ARPHRD_IPDDP:
		return "IPDDP"
	case unix.ARPHRD_IPGRE:
		return "IPGRE"
	case unix.ARPHRD_IRDA:
		return "IRDA"
	case unix.ARPHRD_LAPB:
		return "LAPB"
	case unix.ARPHRD_LOCALTLK:
		return "LOCALTLK"
	case unix.ARPHRD_LOOPBACK:
		return "LOOPBACK"
	case unix.ARPHRD_MCTP:
		return "MCTP"
	case unix.ARPHRD_METRICOM:
		return "METRICOM"
	case unix.ARPHRD_NETLINK:
		return "NETLINK"
	case unix.ARPHRD_NETROM:
		return "NETROM"
	case unix.ARPHRD_NONE:
		return "NONE"
	case unix.ARPHRD_PHONET:
		return "PHONET"
	case unix.ARPHRD_PHONET_PIPE:
		return "PHONET_PIPE"
	case unix.ARPHRD_PIMREG:
		return "PIMREG"
	case unix.ARPHRD_PPP:
		return "PPP"
	case unix.ARPHRD_PRONET:
		return "PRONET"
	case unix.ARPHRD_RAWHDLC:
		return "RAWHDLC"
	case unix.ARPHRD_RAWIP:
		return "RAWIP"
	case unix.ARPHRD_ROSE:
		return "ROSE"
	case unix.ARPHRD_RSRVD:
		return "RSRVD"
	case unix.ARPHRD_SIT:
		return "SIT"
	case unix.ARPHRD_SKIP:
		return "SKIP"
	case unix.ARPHRD_SLIP:
		return "SLIP"
	case unix.ARPHRD_SLIP6:
		return "SLIP6"
	case unix.ARPHRD_TUNNEL:
		return "TUNNEL"
	case unix.ARPHRD_TUNNEL6:
		return "TUNNEL6"
	case unix.ARPHRD_VOID:
		return "VOID"
	case unix.ARPHRD_VSOCKMON:
		return "VSOCKMON"
	case unix.ARPHRD_X25:
		return "X25"

	default:
		return "UNKNOWN"
	}
}
