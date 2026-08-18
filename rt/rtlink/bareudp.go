// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// BareUDP is a generic layer-3 encapsulation device using UDP tunnels.
type BareUDP struct {
	Port       uint16
	Type       EtherType
	SrcPortMin uint16
	MultiProto bool
}

// DeviceKind returns `bareudp`.
func (BareUDP) DeviceKind() string {
	return "bareudp"
}

// MarshalAttributes marshals the IFLA_INFO_DATA attributes for a [BareUDP]
// device type.
func (b *BareUDP) MarshalAttributes(attrs *netlink.AttributeEncoder) error {
	attrs.Uint16(unix.IFLA_BAREUDP_PORT, b.Port)
	attrs.Uint16(unix.IFLA_BAREUDP_ETHERTYPE, uint16(b.Type))

	if b.SrcPortMin != 0 {
		attrs.Uint16(unix.IFLA_BAREUDP_SRCPORT_MIN, b.SrcPortMin)
	}

	if b.MultiProto {
		attrs.Flag(unix.IFLA_BAREUDP_MULTIPROTO_MODE)
	}

	return nil
}

// UnmarshalAttributes unmarshals the IFLA_INFO_DATA attributes for a [BareUDP]
// device type.
func (b *BareUDP) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_BAREUDP_PORT:
			b.Port = attr.Uint16()
		case unix.IFLA_BAREUDP_ETHERTYPE:
			b.Type = EtherType(attr.Uint16())
		case unix.IFLA_BAREUDP_SRCPORT_MIN:
			b.SrcPortMin = attr.Uint16()
		case unix.IFLA_BAREUDP_MULTIPROTO_MODE:
			b.MultiProto = true
		}
	}

	return nil
}
