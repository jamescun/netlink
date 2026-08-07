// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package route

import (
	"encoding/binary"
	"net/netip"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/route/internal/rtnetlink"

	"golang.org/x/sys/unix"
)

// BareUDP is a [LinkDriver] for a layer-3 encapsulation device over UDP.
type BareUDP struct {
	Port       uint16
	EtherType  uint16
	SrcPortMin uint16
	MultiProto bool
}

// DriverName returns `bareudp`.
func (BareUDP) DriverName() string {
	return "bareudp"
}

// UnmarshalAttributes unmarshals a [BareUDP] found inside [Link.Info.Driver].
func (b *BareUDP) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_BAREUDP_PORT:
			b.Port = attr.Uint16()
		case unix.IFLA_BAREUDP_ETHERTYPE:
			b.EtherType = attr.Uint16()
		case unix.IFLA_BAREUDP_SRCPORT_MIN:
			b.SrcPortMin = attr.Uint16()
		case unix.IFLA_BAREUDP_MULTIPROTO_MODE:
			b.MultiProto = true
		}
	}

	return nil
}

// IPIP is a [LinkDriver] for an IP-in-IP tunnel device.
type IPIP struct {
	Link            uint32
	Local           netip.Addr
	Remote          netip.Addr
	TTL             uint8
	TOS             uint8
	EncapLimit      uint8
	FlowInfo        uint32
	Flags           uint16
	Proto           uint8
	PmtuDisc        uint8
	EncapType       uint16
	EncapFlags      uint16
	EncapSport      uint16
	EncapDport      uint16
	CollectMetadata bool
	Fwmark          uint32
}

// DriverName returns `ipip`.
func (IPIP) DriverName() string {
	return "ipip"
}

// UnmarshalAttributes unmarshals aan [IPIP] found inside [Link.Info.Driver].
func (i *IPIP) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case rtnetlink.IPTUN_LINK:
			i.Link = attr.Uint32()
		case rtnetlink.IPTUN_LOCAL:
			ip, ok := netip.AddrFromSlice(attr.Bytes())
			if ok {
				i.Local = ip
			}
		case rtnetlink.IPTUN_REMOTE:
			ip, ok := netip.AddrFromSlice(attr.Bytes())
			if ok {
				i.Remote = ip
			}
		case rtnetlink.IPTUN_TTL:
			i.TTL = attr.Uint8()
		case rtnetlink.IPTUN_TOS:
			i.TOS = attr.Uint8()
		case rtnetlink.IPTUN_ENCAP_LIMIT:
			i.EncapLimit = attr.Uint8()
		case rtnetlink.IPTUN_FLOWINFO:
			i.FlowInfo = binary.BigEndian.Uint32(attr.Bytes())
		case rtnetlink.IPTUN_FLAGS:
			i.Flags = binary.BigEndian.Uint16(attr.Bytes())
		case rtnetlink.IPTUN_PROTO:
			i.Proto = attr.Uint8()
		case rtnetlink.IPTUN_PMTUDISC:
			i.PmtuDisc = attr.Uint8()
		case rtnetlink.IPTUN_ENCAP_TYPE:
			i.EncapType = attr.Uint16()
		case rtnetlink.IPTUN_ENCAP_FLAGS:
			i.EncapFlags = attr.Uint16()
		case rtnetlink.IPTUN_ENCAP_SPORT:
			i.EncapSport = binary.BigEndian.Uint16(attr.Bytes())
		case rtnetlink.IPTUN_ENCAP_DPORT:
			i.EncapDport = binary.BigEndian.Uint16(attr.Bytes())
		case rtnetlink.IPTUN_COLLECT_METADATA:
			i.CollectMetadata = true
		case rtnetlink.IPTUN_FWMARK:
			i.Fwmark = attr.Uint32()
		}
	}

	return nil
}
