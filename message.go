// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
)

// BinaryUnmarshaler is an extension of [encoding.BinaryUnmarshaler] interface
// to declare how many bytes are required to satisfy UnmarshalBinary method.
//
// It otherwise functions exactly the same.
type BinaryUnmarshaler interface {
	encoding.BinaryUnmarshaler

	// Len returns the number of bytes needed from the message to satisfy the
	// call to [BinaryUnmarshaler.UnmarshalBinary].
	Len() int
}

type discard struct {
	length int
}

// Discard returns a [BinaryUnmarshaler], for use with a [MessageDecoder] that
// simply discards a fixed-number of bytes from a message.
func Discard(length int) BinaryUnmarshaler {
	return &discard{length: length}
}

func (d *discard) Len() int                       { return d.length }
func (d *discard) UnmarshalBinary(_ []byte) error { return nil }

// Message is a Netlink message, containing the header and payload.
//
// It is wrapped by [MessageEncoder] and [MessageDecoder] for writing and
// reading messages respectively.
type Message interface {
	// Header returns the Netlink message header.
	Header() MessageHeader

	// Data returns the contents of the Netlink messages body, without the
	// header.
	Data() []byte
}

// MessageDecoder is used to read the intermediate family-specific headers and
// attributes of a Netlink message.
type MessageDecoder interface {
	Message

	// Code returns the status code contained within DONE and ERROR type
	// messages.
	//
	// If not one of these message types, zero is returned.
	Code() int

	// Read arbitrary bytes from the message.
	//
	// Reads will not automatically be aligned.
	Read([]byte) (int, error)

	// Unmarshal unmarshal the Netlink message attributes to a type
	// implementing [AttributeUnmarshaler].
	//
	// If the message contains any intermediate family-specific headers, these
	// should be read first using [MessageDecoder.Read] or
	// [MessageDecoder.UnmarshalBytes].
	Unmarshal(AttributeUnmarshaler) error

	// UnmarshalBytes unmarshals arbitrary bytes from the message to a type
	// implementing [BinaryUnmarshaler].
	//
	// If there are not enough bytes in the message to satisfy the read, an
	// error is returned.
	//
	// Reads will not automatically be aligned, consider using [Align].
	UnmarshalBytes(BinaryUnmarshaler) error
}

type messageDecoder struct {
	hdr MessageHeader

	i    int
	buf  []byte
	code int
}

// NewMessageDecoder unmarshals a single Netlink message from bytes, returning
// a [MessageDecoder] for reading it's contents, including ERROR and DONE
// message types.
//
// If a DONE or ERROR message type, the status code will be read and returned
// by [MessageDecoder.Code].
//
// If an ERROR message type, it will automatically be unmarshaled to the
// [Error] type and returned. The caller must check if the status code is zero
// to check for acknowledgement messages, either by checking the code directly
// or using the [IsACK] function.
//
// It will ignore any additional bytes it is given.
func NewMessageDecoder(b []byte) (MessageDecoder, error) {
	b, _, err := cutMessage(b)
	if err != nil {
		return nil, err
	}

	msg := &messageDecoder{
		buf: b[headerLen:],
	}

	err = msg.hdr.UnmarshalBinary(b)
	if err != nil {
		return nil, fmt.Errorf("header: %w", err)
	}

	if msg.hdr.Length > len(b) {
		return nil, fmt.Errorf("needed %d bytes, got %d", msg.hdr.Length, len(b))
	}

	if msg.hdr.Type == ERROR {
		errMsg := &Error{}
		err := errMsg.UnmarshalBinary(b)
		if err != nil {
			// error reading the error message.
			return nil, fmt.Errorf("error message: %w", err)
		}

		return nil, errMsg
	}

	if msg.hdr.Type == DONE {
		msg.code, msg.buf, err = readErrorCode(msg.buf)
		if err != nil {
			return nil, fmt.Errorf("code: %w", err)
		}
	}

	return msg, nil
}

func (md *messageDecoder) Header() MessageHeader { return md.hdr }
func (md *messageDecoder) Data() []byte          { return md.buf }

func (md *messageDecoder) Code() int {
	return md.code
}

func (md *messageDecoder) Read(b []byte) (int, error) {
	if md.i >= len(md.buf) {
		return 0, io.EOF
	}

	n := copy(b, md.buf[md.i:])
	md.i += n

	return n, nil
}

func (md *messageDecoder) Unmarshal(dst AttributeUnmarshaler) error {
	if dst == nil {
		return fmt.Errorf("AttributeUnmarshaler is nil")
	}

	err := dst.UnmarshalAttributes(&AttributeDecoder{buf: md.buf[md.i:]})
	if err != nil {
		return err
	}

	return nil
}

func (md *messageDecoder) UnmarshalBytes(dst BinaryUnmarshaler) error {
	if dst == nil {
		return fmt.Errorf("BinaryUnmarshaler is nil")
	}

	length := dst.Len()
	if length < 0 {
		return fmt.Errorf("cannot read negative bytes, got %d", length)
	}

	if (md.i + length) >= len(md.buf) {
		return fmt.Errorf("needed %d bytes, got %d", length, len(md.buf)-md.i)
	}

	err := dst.UnmarshalBinary(md.buf[md.i : md.i+length])
	if err != nil {
		return err
	}

	md.i += length

	return nil
}

// MessageEncoder is used to build the intermediate family-specific headers and
// attributes of a Netlink message.
type MessageEncoder interface {
	Message

	// SetHeader configures the message type and configuration flags of the
	// message being encoded.
	SetHeader(msgType, flags uint16)

	// Marshal is called to marshal the Netlink message attributes from a type
	// implementing [AttributeMarshaler].
	//
	// If the message contains any intermediate family-specific headers, these
	// should be written first using [MessageEncoder.Write] or
	// [MessageEncoder.MarshalBytes].
	Marshal(AttributeMarshaler) error

	// MarshalBytes marshals arbitrary bytes to the message from types
	// implementing [encoding.BinaryMarshaler].
	//
	// Writes will not automatically be aligned, consider using [Pad].
	MarshalBytes(encoding.BinaryMarshaler) error

	// Write arbitrary bytes to the message.
	//
	// Written bytes will not automatically be aligned, consider using [Pad].
	Write([]byte) (int, error)
}

type messageEncoder struct {
	hdr MessageHeader

	buf []byte
}

func (me *messageEncoder) Header() MessageHeader { return me.hdr }
func (me *messageEncoder) Data() []byte          { return me.buf }

func (me *messageEncoder) SetHeader(msgType, flags uint16) {
	me.hdr.Type = msgType
	me.hdr.Flags = flags
}

func (me *messageEncoder) Marshal(src AttributeMarshaler) error {
	if src == nil {
		return fmt.Errorf("AttributeMarshaler is nil")
	}

	attrs := &AttributeEncoder{buf: me.buf}

	err := src.MarshalAttributes(attrs)
	if err != nil {
		return err
	}

	me.buf = attrs.buf

	return nil
}

func (me *messageEncoder) MarshalBytes(src encoding.BinaryMarshaler) error {
	if src == nil {
		return fmt.Errorf("encoding.BinaryMarshaler is nil")
	}

	if src, ok := src.(encoding.BinaryAppender); ok {
		buf, err := src.AppendBinary(me.buf)
		if err != nil {
			return err
		}

		me.buf = buf
		return nil
	}

	b, err := src.MarshalBinary()
	if err != nil {
		return err
	}

	me.buf = append(me.buf, b...)
	return nil
}

func (me *messageEncoder) Write(b []byte) (int, error) {
	me.buf = append(me.buf, b...)
	return len(b), nil
}

// MessageReader is an iterator that will yield for each Netlink message
// contained within bytes, until there are no more messages or an error occurs.
type MessageReader struct {
	buf []byte
	err error
}

// NewMessageReader initializes a new [MessageReader] iterator that will yield
// for each Netlink message contained within bytes, until there are no more
// messages or an error occurs.
func NewMessageReader(buf []byte) *MessageReader {
	return &MessageReader{
		buf: buf,
	}
}

// Err returns the last error encountered while reading messages, if any.
func (mr *MessageReader) Err() error {
	return mr.err
}

// Each is an [iter.Seq2] iterator, that will yield for each message contained
// for the bytes given to [MessageReader], and well as the logical message
// number. It will yield [DONE] and [ERROR] messages.
//
// Each message will contain a slice of the original bytes, so message reading
// must be completed before those bytes are reused.
//
// If an error occurs, it will be returned by [MessageReader.Err].
func (mr *MessageReader) Each(yield func(int, MessageDecoder) bool) {
	if mr.err != nil {
		// message reader invalidated by previous error.
		return
	}

	// work on our own slice of bytes, to support multiple iterations.
	buf := mr.buf

	i := 0
	for len(buf) > headerLen {
		msg, err := NewMessageDecoder(buf)
		if err != nil {
			mr.err = err
			break
		}

		if !yield(i, msg) {
			break
		}

		i++
		buf = buf[Align(msg.Header().Length):]
	}
}

// cutMessage splits bytes into the portion containing a Netlink message header
// and body, then any remaining bytes, or an error; based on it's uint32
// length prefix.
func cutMessage(b []byte) (msg, after []byte, err error) {
	if len(b) < headerLen {
		err = fmt.Errorf("needed at least %d bytes, got %d", headerLen, len(b))
		return
	}

	length := int(binary.NativeEndian.Uint32(b))
	if len(b) < length {
		err = fmt.Errorf("needed %d bytes, got %d", length, len(b))
	}

	msg = b[:length]
	after = b[length:]
	return
}
