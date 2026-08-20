// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package rtroute implements interacting with the systems network routing
// through the Netlink rtnetlink subsystem.
//
// References:
//   - rtnetlink(7)
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-route.html
package rtroute

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// CacheInfo contains cache information for a [Route].
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-route.html#rt-route-definition-rta-cacheinfo
type CacheInfo struct {
	Clntref uint32
	LastUse uint32
	Expires uint32
	Error   uint32
	Used    uint32

	// only set when peer info is available.
	Id    uint32
	Ts    uint32
	Tsage uint32
}

// UnmarshalBinary unmarshals the rta_cacheinfo struct from bytes.
func (c *CacheInfo) UnmarshalBinary(b []byte) error {
	if len(b) < 20 {
		return fmt.Errorf("expected at least 20 bytes, got %d", len(b))
	}

	c.Clntref = binary.NativeEndian.Uint32(b)
	c.LastUse = binary.NativeEndian.Uint32(b[4:])
	c.Expires = binary.NativeEndian.Uint32(b[8:])
	c.Error = binary.NativeEndian.Uint32(b[12:])
	c.Used = binary.NativeEndian.Uint32(b[16:])

	if len(b) >= 32 {
		c.Id = binary.NativeEndian.Uint32(b[20:])
		c.Ts = binary.NativeEndian.Uint32(b[24:])
		c.Tsage = binary.NativeEndian.Uint32(b[28:])

	}

	return nil
}

// Features are contained within [Metrics].
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
type Features uint32

// Constants for [Features].
const (
	ECH Features = 1 << iota
	SACK
	TIMESTAMP
	ALLFRAG
	TCP_USEC_TS
)

// featureNames is bit shifted through by [Features.String] to build a
// stringified representation.
var featureNames = []string{
	"ECH",
	"SACK",
	"TIMESTAMP",
	"ALLFRAG",
	"TCP_USEC_TS",
}

func (f Features) String() string {
	if f == 0 {
		return "NONE"
	}

	var s strings.Builder

	for i, name := range featureNames {
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

// MfcStats not sure what this is yet.
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
type MfcStats struct {
	Packets uint64
	Bytes   uint64
	WrongIf uint64
}

// UnmarshalBinary unmarshals the rta_mfc_stats struct from bytes.
func (m *MfcStats) UnmarshalBinary(b []byte) error {
	if len(b) != 24 {
		return fmt.Errorf("expected 24 bytes, got %d", len(b))
	}

	m.Packets = binary.NativeEndian.Uint64(b)
	m.Bytes = binary.NativeEndian.Uint64(b[8:])
	m.WrongIf = binary.NativeEndian.Uint64(b[16:])

	return nil
}

// NextHop contains parameters for the next multipath hop.
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
type NextHop struct {
	Len     uint16
	Flags   uint8
	Hops    uint8
	Index   int
	Gateway netip.Addr
}

// Route contains a network route within one of the systems routing tables.
//
// References:
//   - rtnetlink(7)
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-route.html#rt-route-attribute-set-route-attrs
type Route struct {
	Family       Family
	DstLen       uint8
	SrcLen       uint8
	Table        Table
	Proto        Proto
	Scope        Scope
	Type         Type
	Flags        Flags
	Dst          netip.Addr
	Src          netip.Addr
	Iif          int
	Oif          int
	Gateway      netip.Addr
	Priority     uint32
	PrefSrc      netip.Addr
	Metrics      Metrics
	Multipath    []*NextHop
	Flow         uint32
	CacheInfo    *CacheInfo
	Mark         uint32
	MfcStats     *MfcStats
	Via          *Via
	NewDst       netip.Addr
	Pref         uint8
	EncapType    EncapType
	Uid          uint32
	TtlPropagate uint8
	IPProto      uint8
	Sport        uint16
	Dport        uint16
}

// UnmarshalNetlink unmarshals a single [Route] from a Netlink message.
func (r *Route) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	info := &RtMsg{}
	err := msg.UnmarshalBytes(info)
	if err != nil {
		return fmt.Errorf("rtmsg: %w", err)
	}

	r.Family = info.Family
	r.DstLen = info.DstLen
	r.SrcLen = info.SrcLen
	r.Proto = info.Proto
	r.Scope = info.Scope
	r.Type = info.Type

	err = msg.Unmarshal(r)
	if err != nil {
		return err
	}

	// if table has not been set by the uint32 attribute, take the less
	// expressive uint8 from rtmsg.
	if r.Table == 0 {
		r.Table = Table(info.Table)
	}

	return nil
}

// UnmarshalAttributes unmarshals a [Route] attributes from a Netlink message.
//
// It assumes the [RtMsg] header has already been read and values set on the
// [Route].
func (r *Route) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.RTA_DST:
			r.Dst, _ = attr.Addr()
		case unix.RTA_SRC:
			r.Src, _ = attr.Addr()
		case unix.RTA_IIF:
			r.Iif = int(attr.Uint32())
		case unix.RTA_OIF:
			r.Oif = int(attr.Uint32())
		case unix.RTA_GATEWAY:
			r.Gateway, _ = attr.Addr()
		case unix.RTA_PRIORITY:
			r.Priority = attr.Uint32()
		case unix.RTA_PREFSRC:
			r.PrefSrc, _ = attr.Addr()
		case unix.RTA_METRICS:
			err := attr.Unmarshal(&r.Metrics)
			if err != nil {
				return fmt.Errorf("metrics: %w", err)
			}
		case unix.RTA_FLOW:
			r.Flow = attr.Uint32()
		case unix.RTA_CACHEINFO:
			r.CacheInfo = new(CacheInfo)
			err := attr.UnmarshalBytes(r.CacheInfo)
			if err != nil {
				return fmt.Errorf("cache info: %w", err)
			}
		case unix.RTA_TABLE:
			r.Table = Table(attr.Uint32())
		case unix.RTA_MARK:
			r.Mark = attr.Uint32()
		case unix.RTA_MFC_STATS:
			r.MfcStats = new(MfcStats)
			err := attr.UnmarshalBytes(r.MfcStats)
			if err != nil {
				return fmt.Errorf("mfc stats: %w", err)
			}
		case unix.RTA_VIA:
			r.Via = new(Via)
			err := attr.UnmarshalBytes(r.Via)
			if err != nil {
				return fmt.Errorf("via: %w", err)
			}
		case unix.RTA_NEWDST:
			r.NewDst, _ = attr.Addr()
		case unix.RTA_PREF:
			r.Pref = attr.Uint8()
		case unix.RTA_ENCAP_TYPE:
			r.EncapType = EncapType(attr.Uint16())
		case unix.RTA_UID:
			r.Uid = attr.Uint32()
		case unix.RTA_TTL_PROPAGATE:
			r.TtlPropagate = attr.Uint8()
		case unix.RTA_IP_PROTO:
			r.IPProto = attr.Uint8()
		case unix.RTA_SPORT:
			r.Sport = attr.Uint16()
		case unix.RTA_DPORT:
			r.Dport = attr.Uint16()
		}
	}

	return nil
}

// Routes contains a list of [Route] read from a series of Netlink messages.
type Routes []*Route

// UnmarshalNetlink unmarshals one-or-more Netlink messages containing a
// [Route] and appends it to itself.
func (rs *Routes) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	route := &Route{}

	err := route.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	*rs = append(*rs, route)
	return nil
}

// Metrics contains statistics for a [Route].
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-route.html#metrics
type Metrics struct {
	Lock             uint32
	MTU              uint32
	Window           uint32
	Rtt              uint32
	Rttvar           uint32
	Ssthresh         uint32
	Cwnd             uint32
	Advmss           uint32
	Reordering       uint32
	HopLimit         uint32
	InitCwnd         uint32
	Features         Features
	RtoMin           uint32
	InitRwnd         uint32
	QuickAck         uint32
	CcAlgo           string
	FastOpenNoCookie uint32
}

// UnmarshalAttributes unmarshals the metrics contained within a [Route].
func (m *Metrics) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.RTAX_LOCK:
			m.Lock = attr.Uint32()
		case unix.RTAX_MTU:
			m.MTU = attr.Uint32()
		case unix.RTAX_WINDOW:
			m.Window = attr.Uint32()
		case unix.RTAX_RTT:
			m.Rtt = attr.Uint32()
		case unix.RTAX_RTTVAR:
			m.Rttvar = attr.Uint32()
		case unix.RTAX_SSTHRESH:
			m.Ssthresh = attr.Uint32()
		case unix.RTAX_CWND:
			m.Cwnd = attr.Uint32()
		case unix.RTAX_ADVMSS:
			m.Advmss = attr.Uint32()
		case unix.RTAX_REORDERING:
			m.Reordering = attr.Uint32()
		case unix.RTAX_HOPLIMIT:
			m.HopLimit = attr.Uint32()
		case unix.RTAX_INITCWND:
			m.InitCwnd = attr.Uint32()
		case unix.RTAX_FEATURES:
			m.Features = Features(attr.Uint32())
		case unix.RTAX_RTO_MIN:
			m.RtoMin = attr.Uint32()
		case unix.RTAX_INITRWND:
			m.InitRwnd = attr.Uint32()
		case unix.RTAX_QUICKACK:
			m.QuickAck = attr.Uint32()
		case unix.RTAX_CC_ALGO:
			m.CcAlgo = attr.String()
		case unix.RTAX_FASTOPEN_NO_COOKIE:
			m.FastOpenNoCookie = attr.Uint32()
		}
	}

	return nil
}

// Table configures which routing table a [Route] is part of.
//
// References:
//   - linux/include/uapi/linux/rtnetlink.h
type Table uint32

// Constants for [Table].
const (
	COMPAT    Table = 252
	DEFAULT   Table = 253
	MAIN      Table = 254
	LOCALHOST Table = 255
)

func (t Table) String() string {
	switch t {
	case COMPAT:
		return "COMPAT"
	case DEFAULT:
		return "DEFAULT"
	case MAIN:
		return "MAIN"
	case LOCALHOST:
		return "LOCALHOST"

	default:
		return strconv.Itoa(int(t))
	}
}

// Via configures a gateway in a different route family.
//
// References:
//   - rtnetlink(7)
//   - linux/include/uapi/linux/rtnetlink.h
type Via struct {
	Family uint16
	Addr   netip.Addr
}

// UnmarshalBinary unmarshals the rtvia struct from bytes.
func (v *Via) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("expected at least 2 bytes, got %d", len(b))
	}

	v.Family = binary.NativeEndian.Uint16(b)
	v.Addr, _ = netip.AddrFromSlice(b[2:])

	return nil
}
