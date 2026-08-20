// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtroute

import (
	"encoding/binary"
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// EncapType is the encapsulation type configured for a [Route].
//
// References:
//   - linux/includes/uapi/linux/lw_tunnel.h
type EncapType uint16

// Constants for [EncapType].
const (
	MPLS EncapType = 1 + iota
	IP
	ILA
	IP6
	SEG6
	BPF
	SEG6_LOCAL
	RPL
	IOAM6
	XFRM
)

func (e EncapType) String() string {
	switch e {
	case 0:
		return "UNSPEC"
	case MPLS:
		return "MPLS"
	case IP:
		return "IP"
	case ILA:
		return "ILA"
	case IP6:
		return "IP6"
	case SEG6:
		return "SEG6"
	case BPF:
		return "BPF"
	case SEG6_LOCAL:
		return "SEG6_LOCAL"
	case RPL:
		return "RPL"
	case IOAM6:
		return "IOAM6"
	case XFRM:
		return "XFRM"

	default:
		return "UNKNOWN"
	}
}

// Family configures the family of addresses applicable for a [Route].
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

// Flags configure the handling of a [Route].
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
type Flags uint32

// Constants for [Flags].
const (
	NOTIFY         Flags = unix.RTM_F_NOTIFY
	CLONED         Flags = unix.RTM_F_CLONED
	EQUALIZE       Flags = unix.RTM_F_EQUALIZE
	PREFIX         Flags = unix.RTM_F_PREFIX
	LOOKUP_TABLE   Flags = unix.RTM_F_LOOKUP_TABLE
	FIB_MATCH      Flags = unix.RTM_F_FIB_MATCH
	OFFLOAD        Flags = unix.RTM_F_OFFLOAD
	TRAP           Flags = unix.RTM_F_TRAP
	OFFLOAD_FAILED Flags = unix.RTM_F_OFFLOAD_FAILED
)

// flagNames contains a map of the individual flags to their names for
// [Flags.String] to build the stringified representation.
//
// It is implemented as an array to iterate through for consistent ordering.
var flagNames = []struct {
	flag Flags
	str  string
}{
	{NOTIFY, "NOTIFY"},
	{CLONED, "CLONED"},
	{EQUALIZE, "EQUALIZE"},
	{PREFIX, "PREFIX"},
	{LOOKUP_TABLE, "LOOKUP_TABLE"},
	{FIB_MATCH, "FIB_MATCH"},
	{OFFLOAD, "OFFLOAD"},
	{TRAP, "TRAP"},
	{OFFLOAD_FAILED, "OFFLOAD_FAILED"},
}

func (f Flags) String() string {
	var s strings.Builder

	for _, name := range flagNames {
		if f&name.flag != 0 {
			s.WriteByte(' ')
			s.WriteString(name.str)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}

// Proto is the origin of a [Route].
//
// References:
//   - rtnetlink(7)
type Proto uint8

// Constants for [Proto].
const (
	REDIRECT Proto = 1 + iota
	KERNEL
	BOOT
	STATIC

	// these are user defined values for common routing daemons, not
	// interpreted by the kernel itself.
	GATED      Proto = 8
	RA         Proto = 9
	MRT        Proto = 10
	ZEBRA      Proto = 11
	BIRD       Proto = 12
	DNROUTED   Proto = 13
	XORP       Proto = 14
	NTK        Proto = 15
	DHCP       Proto = 16
	MROUTED    Proto = 17
	KEEPALIVED Proto = 18
	BABEL      Proto = 42
	OVN        Proto = 84
	OPENR      Proto = 99
	BGP        Proto = 186
	ISIS       Proto = 187
	OSPF       Proto = 188
	RIP        Proto = 189
	EIGRP      Proto = 192
)

func (p Proto) String() string {
	switch p {
	case 0:
		return "UNSPEC"
	case REDIRECT:
		return "REDIRECT"
	case KERNEL:
		return "KERNEL"
	case BOOT:
		return "BOOT"
	case STATIC:
		return "STATIC"

	case GATED:
		return "GATED"
	case RA:
		return "RA"
	case MRT:
		return "MRT"
	case ZEBRA:
		return "ZEBRA"
	case BIRD:
		return "BIRD"
	case DNROUTED:
		return "DNROUTED"
	case XORP:
		return "XORP"
	case NTK:
		return "NTK"
	case DHCP:
		return "DHCP"
	case MROUTED:
		return "MROUTED"
	case KEEPALIVED:
		return "KEEPALIVED"
	case BABEL:
		return "BABEL"
	case OVN:
		return "OVN"
	case OPENR:
		return "OPENR"
	case BGP:
		return "BGP"
	case ISIS:
		return "ISIS"
	case OSPF:
		return "OSPF"
	case RIP:
		return "RIP"
	case EIGRP:
		return "EIGRP"

	default:
		return "UNKNOWN"
	}
}

// RtMsg is the fixed-length header before [Route] messages.
//
// References:
//   - rtnetlink(7)
type RtMsg struct {
	Family Family
	DstLen uint8
	SrcLen uint8
	TOS    uint8
	Table  uint8
	Proto  Proto
	Scope  Scope
	Type   Type
	Flags  Flags
}

// Len returns the fixed-length of the [RtMsg] header.
func (RtMsg) Len() int { return 12 }

// AppendBinary appends an [RtMsg] to bytes using the host byteorder.
func (r RtMsg) AppendBinary(b []byte) ([]byte, error) {
	b = append(
		b,
		uint8(r.Family), r.DstLen, r.SrcLen, r.TOS, r.Table, uint8(r.Proto),
		uint8(r.Scope), uint8(r.Type),
	)

	b = binary.NativeEndian.AppendUint32(b, uint32(r.Flags))

	return b, nil
}

// MarshalBinary marshals an [RtMsg] to bytes using the host byteorder.
func (r RtMsg) MarshalBinary() ([]byte, error) {
	return r.AppendBinary(make([]byte, 0, 12))
}

// UnmarshalBinary unmarshals an [RtMsg] from bytes using the host byteorder.
//
// It will ignore any additional bytes it is given.
func (r *RtMsg) UnmarshalBinary(b []byte) error {
	if len(b) < 12 {
		return fmt.Errorf("expected 12 bytes, got %d", len(b))
	}

	r.Family = Family(b[0])
	r.DstLen = b[1]
	r.SrcLen = b[2]
	r.TOS = b[3]
	r.Table = b[4]
	r.Proto = Proto(b[5])
	r.Scope = Scope(b[6])
	r.Type = Type(b[7])
	r.Flags = Flags(binary.NativeEndian.Uint32(b[8:]))

	return nil
}

// String returns a string representation of the attribute header for
// debugging.
func (r RtMsg) String() string {
	return fmt.Sprintf(
		";; ->>RTMSG<<- family: %s, dst: %d, src: %d, tos: %d\n;; table: %d, proto: %s, scope: %s, type: %s\n;; flags: %d\n", //nolint
		r.Family, r.DstLen, r.SrcLen, r.TOS, r.Table, r.Proto, r.Scope, r.Type, r.Flags,
	)
}

// Scope configures the network scope for an [addr.Addr] or [Route].
type Scope uint8

// Constants for [Scope].
const (
	UNIVERSE Scope = unix.RT_SCOPE_UNIVERSE
	SITE     Scope = unix.RT_SCOPE_SITE
	LINK     Scope = unix.RT_SCOPE_LINK
	HOST     Scope = unix.RT_SCOPE_HOST
	NOWHERE  Scope = unix.RT_SCOPE_NOWHERE
)

func (s Scope) String() string {
	switch s {
	case UNIVERSE:
		return "UNIVERSE"
	case SITE:
		return "SITE"
	case LINK:
		return "LINK"
	case HOST:
		return "HOST"
	case NOWHERE:
		return "NOWHERE"

	default:
		return "UNKNOWN"
	}
}

// Type is the type of [Route].
//
// References:
//   - rtnetlink(7)
//   - linux/include/uapi/linux/rtnetlink.h
type Type uint8

// Constants for [Type].
const (
	UNICAST Type = 1 + iota
	LOCAL
	BROADCAST
	ANYCAST
	MULTICAST
	BLACKHOLE
	UNREACHABLE
	PROHIBIT
	THROW
	NAT
	XRESOLVE
)

func (t Type) String() string {
	switch t {
	case 0:
		return "UNSPEC"
	case UNICAST:
		return "UNICAST"
	case LOCAL:
		return "LOCAL"
	case BROADCAST:
		return "BROADCAST"
	case ANYCAST:
		return "ANYCAST"
	case MULTICAST:
		return "MULTICAST"
	case BLACKHOLE:
		return "BLACKHOLE"
	case UNREACHABLE:
		return "UNREACHABLE"
	case PROHIBIT:
		return "PROHIBIT"
	case THROW:
		return "THROW"
	case NAT:
		return "NAT"
	case XRESOLVE:
		return "XRESOLVE"

	default:
		return "UNKNOWN"
	}
}
