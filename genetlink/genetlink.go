// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package genetlink implements the Linux Kernel Generic Netlink protocol, for
// interacting with the systems network stack and related subsystems.
//
// This package contains a low-level implementation of the protocol, you should
// use a higher-level abstraction contained within a subpackage.
//
// References:
//   - linux/include/uapi/linux/netlink.h
//   - linux/include/uapi/linux/genetlink.h
//   - https://www.kernel.org/doc/html/latest/netlink/specs/nlctrl.html
package genetlink

import (
	"encoding/binary"
	"fmt"

	"go.jamescun.com/netlink"
)

// Header is the fixed-length preamble before each Generic Netlink message that
// describes the command.
//
// References:
//   - linux/include/uapi/linux/genetlink.h
type Header struct {
	netlink.MessageHeader

	Cmd      uint8
	Version  uint8
	Reserved uint16
}

// Len returns the fixed-length of the Generic Netlink header.
func (Header) Len() int { return 4 }

// AppendBinary appends a Generic Netlink header to bytes, in the host
// byteorder.
func (h Header) AppendBinary(b []byte) ([]byte, error) {
	b = append(b, h.Cmd, h.Version)
	b = binary.NativeEndian.AppendUint16(b, h.Reserved)
	return b, nil
}

// MarshalBinary marshals a Generic Netlink header to bytes, in the host
// byteorder.
func (h Header) MarshalBinary() ([]byte, error) {
	return h.AppendBinary(make([]byte, 0, 4))
}

// UnmarshalBinary unmarshals a Generic Netlink header from bytes using the
// host byteorder.
//
// It will ignore any additional bytes it is given.
func (h *Header) UnmarshalBinary(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("expected 4 bytes, got %d", len(b))
	}

	h.Cmd = b[0]
	h.Version = b[1]
	h.Reserved = binary.NativeEndian.Uint16(b[2:])

	return nil
}

// String returns a string representation of the attribute header for
// debugging.
func (h Header) String() string {
	return fmt.Sprintf(
		";; ->>GENETLINK<<- cmd: %d, version: %d\n",
		h.Cmd, h.Version,
	)
}
