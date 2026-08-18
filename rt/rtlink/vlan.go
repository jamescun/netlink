// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"encoding/binary"
	"fmt"
	"strings"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// VLAN is the [InfoData] for a 802.1Q Virtual LAN device.
//
// References:
//   - linux/net/8021q/vlan_netlink.c
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-link.html#vlan-protocols
type VLAN struct {
	ID       uint16
	Flags    VLANFlags
	Egress   []*VLANQOS
	Ingress  []*VLANQOS
	Protocol uint16
}

// DeviceKind returns `vlan`.
func (VLAN) DeviceKind() string {
	return "vlan"
}

// MarshalAttributes marshals the IFLA_INFO_DATA attributes for a [VLAN] device
// type.
func (VLAN) MarshalAttributes(_ *netlink.AttributeDecoder) error {
	return fmt.Errorf("VLAN devices cannot be configured yet")
}

// UnmarshalAttributes unmarshals the attributes for a [VLAN] from inside a
// [Link.Info].
func (v *VLAN) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_VLAN_ID:
			v.ID = attr.Uint16()

		case unix.IFLA_VLAN_FLAGS:
			// IFLA_VLAN_FLAGS is actually a struct containing the flags and a
			// change mask, however the latter is only used during
			// configuration, so it can safely be ignored here.
			v.Flags = VLANFlags(attr.Uint32())

		case unix.IFLA_VLAN_EGRESS_QOS:
			for nested := range attr.Each {
				qos := new(VLANQOS)
				err := nested.UnmarshalBytes(qos)
				if err != nil {
					return fmt.Errorf("egress qos: %w", err)
				}

				v.Egress = append(v.Egress, qos)
			}

		case unix.IFLA_VLAN_INGRESS_QOS:
			for nested := range attr.Each {
				qos := new(VLANQOS)
				err := nested.UnmarshalBytes(qos)
				if err != nil {
					return fmt.Errorf("ingress qos: %w", err)
				}

				v.Ingress = append(v.Ingress, qos)
			}

		case unix.IFLA_VLAN_PROTOCOL:
			v.Protocol = attr.Uint16()
		}
	}

	return nil
}

// VLANFlags configures a [VLAN] device.
//
// References:
//   - linux/include/uapi/linux/if_vlan.h
type VLANFlags uint32

// Constants for [VLANFlags].
const (
	REORDER_HDR VLANFlags = 1 << iota
	GVRP
	LOOSE_BINDING
	MVRP
	BRIDGE_BINDING
)

// vlanFlagNames is bit shifted through by [VLANFlags.String] to build a
// stringified representation.
var vlanFlagNames = []string{
	"REORDER_HDR",
	"GVRP",
	"LOOSE_BINDING",
	"MVRP",
	"BRIDGE_BINDING",
}

func (v VLANFlags) String() string {
	if v == 0 {
		return "NONE"
	}

	var s strings.Builder

	for i, name := range vlanFlagNames {
		if v&(1<<i) != 0 {
			s.WriteByte(' ')
			s.WriteString(name)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}

// VLANQOS configures the Quality of Service for a [VLAN] on egress or ingress.
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/rt-link.html#rt-link-definition-ifla-vlan-qos-mapping
type VLANQOS struct {
	From uint32
	To   uint32
}

// Len returns the fixed-length of [VLANQOS].
func (VLANQOS) Len() int { return 8 }

// AppendBinary appends the VLAN QOS struct to bytes, using the host byteorder.
func (v VLANQOS) AppendBinary(b []byte) ([]byte, error) {
	b = binary.NativeEndian.AppendUint32(b, v.From)
	b = binary.NativeEndian.AppendUint32(b, v.To)
	return b, nil
}

// MarshalBinary marshals the VLAN QOS struct to bytes, using the host
// byteorder.
func (v VLANQOS) MarshalBinary() ([]byte, error) {
	return v.AppendBinary(make([]byte, 0, 8))
}

// UnmarshalBinary unmarshals the VLAN QOS struct from bytes.
func (v *VLANQOS) UnmarshalBinary(b []byte) error {
	if len(b) != 8 {
		return fmt.Errorf("expected 8 bytes, got %d", len(b))
	}

	v.From = binary.NativeEndian.Uint32(b)
	v.To = binary.NativeEndian.Uint32(b[4:])

	return nil
}
