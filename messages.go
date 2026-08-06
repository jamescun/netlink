// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"fmt"
	"io"
)

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
