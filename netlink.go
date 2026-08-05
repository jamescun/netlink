// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package netlink implements the Linux Kernel Netlink protocol, for
// interacting with the systems network stack and related subsystems.
//
// This package contains a low-level implementation of the protocol, you should
// use a higher-level abstraction contained within a subpackage.
//
// References:
//   - linux/include/uapi/linux/netlink.h
//   - https://www.kernel.org/doc/html/next/userspace-api/netlink/intro.html
//   - https://www.infradead.org/~tgr/libnl/doc/core.html
package netlink

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

const (
	align      = 4  // number of bytes to align messages and attributes to.
	hdrLen     = 16 // length of the message header in bytes.
	attrHdrLen = 4  // length of the attribute header in bytes.
)

// Align return length n as aligned to 4 bytes.
func Align(n int) int {
	return (n + align - 1) & ^(align - 1)
}

// Pad appends up to 4 null bytes to the end of b to align it.
func Pad(b []byte) []byte {
	const pad = "\x00\x00\x00\x00"

	n := Align(len(b)) - len(b)
	return append(b, pad[:n]...)
}

// Family configures the family of the Unix socket used for Netlink messages.
type Family int

// Constants for [Family].
const (
	ROUTE          Family = 0
	UNUSED         Family = 1
	USERSOCK       Family = 2
	FIREWALL       Family = 3
	SOCK_DIAG      Family = 4 //nolint
	NFLOG          Family = 5
	XFRM           Family = 6
	SELINUX        Family = 7
	ISCSI          Family = 8
	AUDIT          Family = 9
	FIB_LOOKUP     Family = 10 //nolint
	CONNECTOR      Family = 11
	NETFILTER      Family = 12
	IP6_FW         Family = 13 //nolint
	DNRTMSG        Family = 14
	KOBJECT_UEVENT Family = 15 //nolint
	GENERIC        Family = 16
	SCSITRANSPORT  Family = 18
	ECRYPTFS       Family = 19
	RDMA           Family = 20
	CRYPTO         Family = 21
	SMC            Family = 22
)

func (f Family) String() string {
	switch f {
	case ROUTE:
		return "ROUTE"
	case UNUSED:
		return "UNUSED"
	case USERSOCK:
		return "USERSOCK"
	case FIREWALL:
		return "FIREWALL"
	case SOCK_DIAG:
		return "SOCK_DIAG"
	case NFLOG:
		return "NFLOG"
	case XFRM:
		return "XFRM"
	case SELINUX:
		return "SELINUX"
	case ISCSI:
		return "ISCSI"
	case AUDIT:
		return "AUDIT"
	case FIB_LOOKUP:
		return "FIB_LOOKUP"
	case CONNECTOR:
		return "CONNECTOR"
	case NETFILTER:
		return "NETFILTER"
	case IP6_FW:
		return "IP6_FW"
	case DNRTMSG:
		return "DNRTMSG"
	case KOBJECT_UEVENT:
		return "KOBJECT_UEVENT"
	case GENERIC:
		return "GENERIC"
	case SCSITRANSPORT:
		return "SCSITRANSPORT"
	case ECRYPTFS:
		return "ECRYPTFS"
	case RDMA:
		return "RDMA"
	case CRYPTO:
		return "CRYPTO"
	case SMC:
		return "SMC"

	default:
		return fmt.Sprintf("Family(%d)", f)
	}
}

// Constants for builtin Netlink message types.
const (
	NOOP    uint16 = 0x01
	ERROR   uint16 = 0x02
	DONE    uint16 = 0x03
	OVERRUN uint16 = 0x04
)

// Flags configures a Netlink message.
type Flags uint16

// Constants for [Flags].
const (
	REQUEST       Flags = 0x01
	MULTI         Flags = 0x02
	ACK           Flags = 0x04
	ECHO          Flags = 0x08
	DUMP_INTR     Flags = 0x10 //nolint
	DUMP_FILTERED Flags = 0x20 //nolint

	// GET messages.
	ROOT   Flags = 0x100
	MATCH  Flags = 0x200
	ATOMIC Flags = 0x400
	DUMP   Flags = (ROOT | MATCH)

	// NEW messages.
	REPLACE Flags = 0x100
	EXCL    Flags = 0x200
	CREATE  Flags = 0x400
	APPEND  Flags = 0x800

	// DELETE messages.
	NONREC Flags = 0x100
	BULK   Flags = 0x200

	// ACK messages.
	CAPPED   Flags = 0x100
	ACK_TLVS Flags = 0x200 //nolint
)

// flagNames is bit shifted through by [Flags.String] to build a stringified
// representation of the common flags.
var flagNames = []string{
	"REQUEST",
	"MULTI",
	"ACK",
	"ECHO",
	"DUMP_INTR",
	"DUMP_FILTERED",
}

func (f Flags) String() string {
	const charset = `0123456789abcdef`

	var s strings.Builder

	// write the per-request flags as a hex digit.
	left := uint8(f >> 8)
	s.WriteByte(charset[left>>4])
	s.WriteByte(charset[left&0x0F])

	// write the named flags, delimited by '|'.
	for i := range 6 {
		if f&(1<<i) != 0 {
			s.WriteByte('|')
			s.WriteString(flagNames[i])
		}
	}

	return s.String()
}

// Constants for builtin Netlink attribute types.
const (
	FLAG         uint16 = 1
	U8           uint16 = 2
	U16          uint16 = 3
	U32          uint16 = 4
	U64          uint16 = 5
	S8           uint16 = 6
	S16          uint16 = 7
	S32          uint16 = 8
	S64          uint16 = 9
	BINARY       uint16 = 10
	STRING       uint16 = 11
	NUL_STRING   uint16 = 12 //nolint
	NESTED       uint16 = 13
	NESTED_ARRAY uint16 = 14 //nolint
	BITFIELD32   uint16 = 15
	SINT         uint16 = 16
	UINT         uint16 = 17

	// Attribute type flags, not used by all families.
	F_NESTED        uint16 = (1 << 15)                     //nolint
	F_NET_BYTEORDER uint16 = (1 << 14)                     //nolint
	F_TYPE_MASK     uint16 = ^(F_NESTED | F_NET_BYTEORDER) //nolint
)

// Header is the fixed-length preamble before each Netlink message that
// describes it's contents.
type Header struct {
	Length int    // including the header itself.
	Type   uint16 // type of message.
	Flags  Flags  // message flags.
	Seq    uint32 // message sequence id.
	Pid    uint32 // socket port id.
}

// AppendBinary marshals the Netlink header, using the host byteorder, and
// appends it to bytes.
//
// If length is not set, it will be set automatically to the header length.
func (h Header) AppendBinary(b []byte) ([]byte, error) {
	if h.Length == 0 {
		h.Length = hdrLen
	}

	if h.Length < hdrLen {
		return nil, fmt.Errorf("header: invalid length, got %d", h.Length)
	} else if h.Length > math.MaxUint32 {
		return nil, fmt.Errorf("header: length exceeds uint32, got %d", h.Length)
	}

	b = binary.NativeEndian.AppendUint32(b, uint32(h.Length))
	b = binary.NativeEndian.AppendUint16(b, h.Type)
	b = binary.NativeEndian.AppendUint16(b, uint16(h.Flags))
	b = binary.NativeEndian.AppendUint32(b, h.Seq)
	b = binary.NativeEndian.AppendUint32(b, h.Pid)

	return b, nil
}

// MarshalBinary marshals the Netlink header, using the host byteorder, as
// bytes.
//
// If length is not set, it will be set automatically to the header length.
func (h Header) MarshalBinary() ([]byte, error) {
	return h.AppendBinary(nil)
}

// UnmarshalBinary unmarshal a Netlink header, using the host byteorder, from
// bytes.
//
// It will ignore any additional bytes it is given.
func (h *Header) UnmarshalBinary(b []byte) error {
	if len(b) < hdrLen {
		return fmt.Errorf("header: expected at least %d bytes, got %d", hdrLen, len(b))
	}

	h.Length = int(binary.NativeEndian.Uint32(b))
	h.Type = binary.NativeEndian.Uint16(b[4:])
	h.Flags = Flags(binary.NativeEndian.Uint16(b[6:]))
	h.Seq = binary.NativeEndian.Uint32(b[8:])
	h.Pid = binary.NativeEndian.Uint32(b[12:])

	if h.Length < hdrLen {
		// length does not include the header itself.
		return fmt.Errorf("header: invalid length, must be at least %d bytes, got %d", hdrLen, h.Length)
	}

	return nil
}

// String returns a string representation of the header for debugging.
func (h Header) String() string {
	return fmt.Sprintf(
		";; ->>NETLINK<<- type: %d, flags: %s\n;; length: %d, seq: %d, pid: %d\n",
		h.Type, h.Flags, h.Length, h.Seq, h.Pid,
	)
}

// Message contains the header, optional intermediate family-specific header,
// and attributes of a Netlink message, in the host byteorder.
type Message struct {
	Header

	buf []byte
}

// AppendBinary marshals the Netlink header, using the host byteorder, and
// appends it to bytes.
//
// If length is not set, it will be set automatically.
func (m *Message) AppendBinary(b []byte) ([]byte, error) {
	if m.Length == 0 {
		m.Length = hdrLen + len(m.buf)
	}

	b, err := m.Header.AppendBinary(b)
	if err != nil {
		return nil, fmt.Errorf("message: %w", err)
	}

	b = append(b, m.buf...)
	b = Pad(b)

	return b, nil
}

// Marshal is called to marshal the Netlink message attributes from a type
// implementing [AttributeMarshaler] to the message.
//
// Intermediate family-specific headers should be written before progressing to
// writing attributes.
func (m *Message) Marshal(am AttributeMarshaler) error {
	if am == nil {
		return fmt.Errorf("message: attributes: received nil AttributeMarshaler")
	}

	attrs := &AttributeWriter{
		buf: m.buf,
	}

	err := am.MarshalAttributes(attrs)
	if err != nil {
		return fmt.Errorf("message: attributes: %w", err)
	}

	m.buf = attrs.buf
	attrs.buf = nil

	return nil
}

// MarshalBinary marshals the Netlink header, using the host byteorder, to
// bytes.
//
// If length is not set, it will be set automatically.
func (m *Message) MarshalBinary() ([]byte, error) {
	return m.AppendBinary(nil)
}

// Read arbitrary bytes from the Netlink message body. This is used to
// intercept intermediate family-specific headers before moving on to the
// attributes that follow.
//
// The bytes read will not automatically be aligned to 4 bytes, consider using
// [Align] to read the correct number of bytes.
func (m *Message) Read(b []byte) (int, error) {
	if len(m.buf) == 0 {
		return 0, io.EOF
	}

	n := copy(b, m.buf)
	m.buf = m.buf[n:]

	return n, nil
}

// Unmarshal is called to unmarshal the Netlink message attributes to a type
// implementing [AttributeUnmarshaler] fromm the message.
//
// Intermediate family-specific headers should be read before progressing to
// attribute handling.
func (m *Message) Unmarshal(au AttributeUnmarshaler) error {
	if au == nil {
		return fmt.Errorf("message: attributes: received nil AttributeUnmarshaler")
	}

	attrs := &AttributeReader{
		buf: m.buf,
	}

	err := au.UnmarshalAttributes(attrs)
	if err != nil {
		return fmt.Errorf("message: attributes: %w", err)
	}

	return nil
}

// UnmarshalBinary unmarshals a Netlink message from bytes, using the host
// byteorder.
//
// It will ignore any additional bytes it is given.
func (m *Message) UnmarshalBinary(b []byte) error {
	err := m.Header.UnmarshalBinary(b)
	if err != nil {
		return fmt.Errorf("message: %w", err)
	}

	if len(b) < Align(m.Length) {
		return fmt.Errorf("message: expected %d bytes, got %d", m.Length, len(b))
	}

	m.buf = b[hdrLen:m.Length]

	return nil
}

// Write arbitrary bytes to the Netlink message body. This is used to
// prepend intermediate family-specific headers before the attributes are
// marshaled.
//
// The bytes written will nto automatically aligned to 4 bytes, consider using
// [Pad] to write the correct number of bytes.
func (m *Message) Write(b []byte) (int, error) {
	m.buf = append(m.buf, b...)
	return len(b), nil
}
