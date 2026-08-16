// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"encoding/binary"
	"fmt"
	"math"
)

// headerLen is the length of the netlink message header in bytes.
const headerLen = 16

// MessageHeader contains the fixed-length preamble at the beginning of a
// Netlink message.
type MessageHeader struct {
	Length int
	Type   uint16
	Flags  uint16
	Seq    uint32
	Pid    uint32
}

// AppendBinary appends the Netlink message header to bytes, using the host
// byteorder.
//
// The total length of the message cannot exceed a uint32.
func (mh MessageHeader) AppendBinary(b []byte) ([]byte, error) {
	if mh.Length < 0 || mh.Length > math.MaxUint32 {
		return nil, fmt.Errorf("message length exceeds uint32, got %d", mh.Length)
	}

	b = binary.NativeEndian.AppendUint32(b, uint32(mh.Length))
	b = binary.NativeEndian.AppendUint16(b, mh.Type)
	b = binary.NativeEndian.AppendUint16(b, mh.Flags)
	b = binary.NativeEndian.AppendUint32(b, mh.Seq)
	b = binary.NativeEndian.AppendUint32(b, mh.Pid)

	return b, nil
}

// MarshalBinary marshals the Netlink message header to bytes, using the host
// byteorder.
//
// The total length of the message cannot exceed a uint32.
func (mh MessageHeader) MarshalBinary() ([]byte, error) {
	return mh.AppendBinary(make([]byte, 0, headerLen))
}

// MarshalNetlink implemented [Marshaler] for writing messages that just
// contain a header and no body.
func (mh MessageHeader) MarshalNetlink(msg MessageEncoder) error {
	msg.SetHeader(mh.Type, mh.Flags)
	return nil
}

// UnmarshalBinary unmarshals the Netlink message header from bytes, using the
// host byteorder.
func (mh *MessageHeader) UnmarshalBinary(b []byte) error {
	if len(b) < headerLen {
		return fmt.Errorf("needed %d bytes, got %d", headerLen, len(b))
	}

	mh.Length = int(binary.NativeEndian.Uint32(b))
	mh.Type = binary.NativeEndian.Uint16(b[4:])
	mh.Flags = binary.NativeEndian.Uint16(b[6:])
	mh.Seq = binary.NativeEndian.Uint32(b[8:])
	mh.Pid = binary.NativeEndian.Uint32(b[12:])

	if mh.Length < headerLen {
		return fmt.Errorf("invalid length, needed at least %d bytes, got %d", headerLen, mh.Length)
	}

	return nil
}

func (mh MessageHeader) String() string {
	return fmt.Sprintf(
		";; ->>NETLINK<<- type: %d, flags: %04x\n;; length: %d, seq: %d, pid: %d\n",
		mh.Type, mh.Flags, mh.Length, mh.Seq, mh.Pid,
	)
}

// cutHeader splits bytes into the portion containing a Netlink header and any
// remaining bytes, or an error; it does not validate the full body length,
func cutHeader(b []byte) (hdr, after []byte, err error) {
	if len(b) < headerLen {
		err = fmt.Errorf("needed at least %d bytes, got %d", headerLen, len(b))
		return
	}

	hdr = b[:headerLen]
	after = b[headerLen:]
	return
}
