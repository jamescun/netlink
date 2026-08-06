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
		return nil, err
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
		return fmt.Errorf("attributes: received nil AttributeMarshaler")
	}

	attrs := &AttributeWriter{
		buf: m.buf,
	}

	err := am.MarshalAttributes(attrs)
	if err != nil {
		return fmt.Errorf("attributes: %w", err)
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
		return fmt.Errorf("attributes: received nil AttributeUnmarshaler")
	}

	attrs := &AttributeReader{
		buf: m.buf,
	}

	err := au.UnmarshalAttributes(attrs)
	if err != nil {
		return fmt.Errorf("attributes: %w", err)
	}

	return nil
}

// UnmarshalBinary unmarshals a Netlink message from bytes, using the host
// byteorder.
//
// The message will contain a slice of the original bytes, so message reading
// must be completed before those bytes are reused.
//
// It will ignore any additional bytes it is given.
func (m *Message) UnmarshalBinary(b []byte) error {
	err := m.Header.UnmarshalBinary(b)
	if err != nil {
		return err
	}

	if len(b) < Align(m.Length) {
		return fmt.Errorf("expected %d bytes, got %d", m.Length, len(b))
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

// MessageReader is used to iterate through one-or-more encoded Netlink
// messages from bytes, using the host native byteorder.
type MessageReader struct {
	buf []byte
	err error
}

// NewMessageReader initializes a new [MessageReader] from bytes containing
// one-or-more messages to read.
func NewMessageReader(buf []byte) *MessageReader {
	return &MessageReader{
		buf: buf,
	}
}

// Err returns the last error encountered while reading messages, if any.
func (mr *MessageReader) Err() error {
	return mr.err
}

// Each is an [iter.Seq2] iterator, that will yield for each [Message]
// contained for the bytes given to [MessageReader], and well as the logical
// message number.
//
// Each message will contain a slice of the original bytes, so message reading
// must be completed before those bytes are reused.
//
// If an error occurs, it will be returned by [MessageReader.Err].
func (mr *MessageReader) Each(yield func(int, *Message) bool) {
	if mr.err != nil {
		// message reader invalidated by previous error.
		return
	}

	// work on our own slice of bytes, to support multiple iterations.
	buf := mr.buf

	i := 0
	for len(buf) > hdrLen {
		msg := &Message{}
		err := msg.UnmarshalBinary(buf)
		if err != nil {
			mr.err = fmt.Errorf("message %d: %w", i, err)
			break
		}

		if !yield(i, msg) {
			break
		}

		i++
		buf = buf[Align(msg.Length):]
	}
}
