// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtaddr

import (
	"encoding/binary"
	"fmt"
	"math"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/rt/rtroute"
)

// IfAddrMsg is the fixed-length header before [Addr] messages.
type IfAddrMsg struct {
	Family    rtroute.Family
	PrefixLen uint8
	Flags     Flags
	Scope     rtroute.Scope
	Link      int
}

// Len returns the fixed-length of the [IfAddrMsg] header.
func (IfAddrMsg) Len() int { return 8 }

// AppendBinary appends an [IfAddrMsg] to bytes using the host byteorder.
func (i IfAddrMsg) AppendBinary(b []byte) ([]byte, error) {
	if i.Link < 0 || i.Link > math.MaxUint32 {
		return nil, fmt.Errorf("index exceeds uint32: %d", i.Link)
	}

	// NOTE(jc): the flags in IfAddrMsg have been superseded by IFA_FLAGS,
	// which is wider and contains more options. The kernel will prefer that
	// when set.

	b = append(b, uint8(i.Family), i.PrefixLen, uint8(i.Flags), uint8(i.Scope)) //nolint
	b = binary.NativeEndian.AppendUint32(b, uint32(i.Link))

	return b, nil
}

// MarshalBinary marshals an [IfAddrMsg] to bytes using the host byteorder.
func (i IfAddrMsg) MarshalBinary() ([]byte, error) {
	return i.AppendBinary(make([]byte, 0, 8))
}

// MarshalNetlink is implement to marshal Netlink messages that only require
// the [IfAddrMsg] header.
func (i IfAddrMsg) MarshalNetlink(msg netlink.MessageEncoder) error {
	err := msg.MarshalBytes(i)
	if err != nil {
		return fmt.Errorf("ifaddrmsg: %w", err)
	}

	return nil
}

// UnmarshalBinary unmarshals an [IfAddrMsg] from bytes using the host
// byteorder.
//
// It will ignore any additional bytes it is given.
func (i *IfAddrMsg) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("expected 8 bytes, got %d", len(b))
	}

	i.Family = rtroute.Family(b[0])
	i.PrefixLen = b[1]
	i.Flags = Flags(b[2])
	i.Scope = rtroute.Scope(b[3])
	i.Link = int(binary.NativeEndian.Uint32(b[4:]))

	return nil
}

// String returns a string representation of the attribute header for
// debugging.
func (i IfAddrMsg) String() string {
	return fmt.Sprintf(
		";; ->>IFADDRMSG<<- family: %s, prefix: %d, scope: %s, link: %d\n;; flags: %s\n",
		i.Family, i.PrefixLen, i.Scope, i.Link, i.Flags,
	)
}
