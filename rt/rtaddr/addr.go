// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package rtaddr implements interacting with the systems network addresses
// through the Netlink rtnetlink subsystem.
//
// References:
//   - rtnetlink(7)
//   - linux/include/uapi/linux/rtnetlink.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-addr.html
package rtaddr

import (
	"encoding/binary"
	"fmt"
	"math"
	"net/netip"
	"strings"
	"time"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/rt/rtroute"

	"golang.org/x/sys/unix"
)

// Addr contains a network address associated with a [Link].
//
// References:
//   - linux/include/uapi/linux/if_addr.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-addr.html#addr-attrs
type Addr struct {
	Family        rtroute.Family
	PrefixLen     uint8
	Scope         rtroute.Scope
	Link          int
	Address       netip.Prefix
	Local         netip.Prefix
	Label         string
	Broadcast     netip.Addr
	Anycast       netip.Addr
	CacheInfo     CacheInfo
	Flags         Flags
	RtPriority    uint32
	TargetNetNsId uint32
	Proto         Proto
}

// UnmarshalAttributes unmarshals the Netlink message attributes for an [Addr].
func (a *Addr) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFA_ADDRESS:
			ip, ok := attr.Addr()
			if ok {
				a.Address = netip.PrefixFrom(ip, int(a.PrefixLen))
			}
		case unix.IFA_LOCAL:
			ip, ok := attr.Addr()
			if ok {
				a.Local = netip.PrefixFrom(ip, int(a.PrefixLen))
			}
		case unix.IFA_LABEL:
			a.Label = attr.String()
		case unix.IFA_BROADCAST:
			a.Broadcast, _ = attr.Addr()
		case unix.IFA_ANYCAST:
			a.Anycast, _ = attr.Addr()
		case unix.IFA_CACHEINFO:
			err := attr.UnmarshalBytes(&a.CacheInfo)
			if err != nil {
				return fmt.Errorf("cache info: %w", err)
			}
		case unix.IFA_FLAGS:
			a.Flags = Flags(attr.Uint32())
		case unix.IFA_RT_PRIORITY:
			a.RtPriority = attr.Uint32()
		case unix.IFA_TARGET_NETNSID:
			a.TargetNetNsId = attr.Uint32()
		case 11: // IFA_PROTO
			a.Proto = Proto(attr.Uint8())
		}
	}

	return nil
}

// UnmarshalNetlink unmarshals a Netlink message containing an [Addr].
func (a *Addr) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	info := IfAddrMsg{}
	err := msg.UnmarshalBytes(&info)
	if err != nil {
		return fmt.Errorf("ifaddrmsg: %w", err)
	}

	a.Family = info.Family
	a.PrefixLen = info.PrefixLen
	a.Scope = info.Scope
	a.Link = info.Link

	err = msg.Unmarshal(a)
	if err != nil {
		return err
	}

	// IFLA_FLAGS is a newer superset of the IfAddrMsg flags, fallback only if
	// that attribute is not set.
	if a.Flags == 0 {
		a.Flags = Flags(info.Flags)
	}

	return nil
}

// Addrs contains a list of [Addr] read from a series of Netlink messages.
type Addrs []*Addr

// FilterFamily filters [Addrs] by family, returning the filtered list.
//
// The original list remains unmodified.
func (as Addrs) FilterFamily(family rtroute.Family) Addrs {
	filtered := Addrs{}

	for _, a := range as {
		if a.Family == family {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

// FilterLink filters [Addrs] by link index, returning the filtered list.
//
// The original list remains unmodified.
func (as Addrs) FilterLink(link int) Addrs {
	filtered := Addrs{}

	for _, a := range as {
		if a.Link == link {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

// UnmarshalNetlink unmarshals one-or-more Netlink messages containing an
// [Addr] and appends it to itself.
func (as *Addrs) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	addr := &Addr{}

	err := addr.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	*as = append(*as, addr)
	return nil
}

// Forever is the value of an address lifetime that represents forever.
const Forever = time.Duration(math.MaxUint32)

// CacheInfo contains information about the lifetime of an [Addr], such as
// for addresses configured through DHCP.
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-addr.html#ifa-cacheinfo
type CacheInfo struct {
	Preferred time.Duration
	Valid     time.Duration
	Cstamp    uint32
	Tstamp    uint32
}

// AppendBinary appends a [CacheInfo] to bytes using the host byteorder.
//
// If either Preferred or Valid is zero, [Forever] will be used.
func (c CacheInfo) AppendBinary(b []byte) ([]byte, error) {
	preferred := uint32(math.MaxUint32)
	if c.Preferred > 0 {
		preferred = uint32(c.Preferred / time.Second) //nolint
	}

	valid := uint32(math.MaxUint32)
	if c.Valid > 0 {
		valid = uint32(c.Valid / time.Second) //nolint
	}

	b = binary.NativeEndian.AppendUint32(b, preferred)
	b = binary.NativeEndian.AppendUint32(b, valid)
	b = binary.NativeEndian.AppendUint32(b, c.Cstamp)
	b = binary.NativeEndian.AppendUint32(b, c.Tstamp)

	return b, nil
}

// MarshalBinary marshals a [CacheInfo] to bytes using the host byteorder.
//
// If either Preferred or Valid is zero, [Forever] will be used.
func (c CacheInfo) MarshalBinary() ([]byte, error) {
	return c.AppendBinary(make([]byte, 0, 16))
}

// UnmarshalBinary unmarshals an [CacheInfo] from bytes using the host
// byteorder.
func (c *CacheInfo) UnmarshalBinary(b []byte) error {
	if len(b) < 16 {
		return fmt.Errorf("needed 16 bytes, got %d", len(b))
	}

	c.Preferred = time.Duration(binary.NativeEndian.Uint32(b)) * time.Second
	c.Valid = time.Duration(binary.NativeEndian.Uint32(b[4:])) * time.Second
	c.Cstamp = binary.NativeEndian.Uint32(b[8:])
	c.Tstamp = binary.NativeEndian.Uint32(b[12:])

	return nil
}

// Flags contains the status of an [Addr], it is an extension of the flags
// already in [IfAddrMsg.Flags].
//
// References:
//   - linux/include/uapi/linux/if_addr.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-addr.html#ifa-flags
type Flags uint32

// Constants for [Flags].
const (
	SECONDARY Flags = 1 << iota
	NO_DAD
	OPTIMISTIC
	DAD_FAILED
	HOME_ADDRESS
	DEPRECATED
	TENTATIVE
	PERMANENT
	MANAGE_TEMP_ADDR
	NO_PREFIX_ROUTE
	MC_AUTO_JOIN
	STABLE_PRIVACY
)

// flagNames is bit shifted through by [Flags.String] to build a stringified
// representation.
var flagNames = []string{
	"SECONDARY",
	"NO_DAD",
	"OPTIMISTIC",
	"DAD_FAILED",
	"HOME_ADDRESS",
	"DEPRECATE",
	"TENTATIVE",
	"PERMANENT",
	"MANAGE_TEMP_ADDR",
	"NO_PREFIX_ROUTE",
	"MC_AUTO_JOIN",
	"STABLE-PRIVACY",
}

func (f Flags) String() string {
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

// Proto is the kernel protocol identifier for an [Addr].
//
// This type only implements the reserved values, other values may be set.
//
// References:
//   - linux/include/uapi/linux/if_addr.h
type Proto uint8

// Constants for [Proto].
const (
	LOOPBACK Proto = 1 + iota
	ROUTER_ANNOUNCEMENT
	LINK_LOCAL
)

func (a Proto) String() string {
	switch a {
	case 0:
		return "NONE"
	case LOOPBACK:
		return "LOOPBACK"
	case ROUTER_ANNOUNCEMENT:
		return "ROUTER_ANNOUNCEMENT"
	case LINK_LOCAL:
		return "LINK_LOCAL"

	default:
		return "UNKNOWN"
	}
}
