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
	"math"
)

// Attribute contains an encoded Netlink attribute to be unmarshaled.
type Attribute struct {
	t   uint16
	buf []byte
}

// Length returns the length of the attribute, including the attribute header,
// in bytes, but excluding any alignment padding.
func (a Attribute) Length() int {
	return attrHdrLen + len(a.buf)
}

// Type returns the type of the attribute from the attribute header.
func (a Attribute) Type() uint16 {
	return a.t
}

// Array is an [iter.Seq] iterator that will yield an [AttributeReader] for
// each nested array item in this attribute.
func (a Attribute) Array(yield func(*AttributeReader) bool) {
	attrs := &AttributeReader{
		buf: a.buf,
	}

	for attr := range attrs.Each {
		if !yield(&AttributeReader{buf: attr.buf}) {
			break
		}
	}
}

// Bytes returns the contents of the attribute as bytes, excluding the
// attribute header or any alignment bytes.
func (a Attribute) Bytes() []byte {
	b := make([]byte, len(a.buf))
	copy(b, a.buf)
	return b
}

// Copy the contents of the attributes bytes into dst, returning how many
// bytes were copied.
//
// Multiple calls to Copy will yield the same bytes.
func (a Attribute) Copy(dst []byte) int {
	return copy(dst, a.buf)
}

// Int8 unmarshal the attribute as an int8.
//
// If not an int8, zero is returned.
func (a Attribute) Int8() int8 {
	return int8(a.Uint8()) //nolint
}

// Int16 unmarshal the attribute as an int16.
//
// If not an int16, zero is returned.
func (a Attribute) Int16() int16 {
	return int16(a.Uint16()) //nolint
}

// Int32 unmarshal the attribute as an int32.
//
// If not an int32, zero is returned.
func (a Attribute) Int32() int32 {
	return int32(a.Uint32()) //nolint
}

// Int64 unmarshal the attribute as an int64.
//
// If not an int64, zero is returned.
func (a Attribute) Int64() int64 {
	return int64(a.Uint64()) //nolint
}

// Nested returns an [AttributeReader] configured with the contents of this
// attribute for handling nested attributes.
//
// You may also want to consider [Attribute.Unmarshal] for types that implement
// the [AttributeUnmarshaler] interface.
func (a Attribute) Nested() *AttributeReader {
	return &AttributeReader{
		buf: a.buf,
	}
}

// String returns the contents of the attribute as a string, excluding the
// null-terminator.
//
// If not a string, the string will contain gibberish.
func (a Attribute) String() string {
	if len(a.buf) == 0 {
		return ""
	}

	buf := a.buf

	// trim null terminator, if it has one, before casting to a string.
	if buf[len(buf)-1] == 0x00 {
		buf = buf[:len(buf)-1]
	}

	return string(buf)
}

// Uint8 unmarshal the attribute as a uint8.
//
// If not a uint8, zero is returned.
func (a Attribute) Uint8() uint8 {
	if len(a.buf) < 1 {
		return 0
	}

	return a.buf[0]
}

// Uint16 unmarshal the attribute as a uint16.
//
// If not a uint16, zero is returned.
func (a Attribute) Uint16() uint16 {
	if len(a.buf) < 2 {
		return 0
	}

	return binary.NativeEndian.Uint16(a.buf)
}

// Uint32 unmarshal the attribute as a uint32.
//
// If not a uint32, zero is returned.
func (a Attribute) Uint32() uint32 {
	if len(a.buf) < 4 {
		return 0
	}

	return binary.NativeEndian.Uint32(a.buf)
}

// Uint64 unmarshal the attribute as a uint64.
//
// If not a uint64, zero is returned.
func (a Attribute) Uint64() uint64 {
	if len(a.buf) < 8 {
		return 0
	}

	return binary.NativeEndian.Uint64(a.buf)
}

// Unmarshal is called to read nested attributes from this attribute using an
// [AttributeUnmarshaler].
//
// You may also want to consider [Attribute.Nested] to iterate through the
// the attributes, if [AttributeUnmarshaler] is not implemented.
func (a Attribute) Unmarshal(au AttributeUnmarshaler) error {
	if au == nil {
		return fmt.Errorf("nested %d: AttributeUnmarshaler is nil", a.t)
	}

	attrs := &AttributeReader{
		buf: a.buf,
	}

	err := au.UnmarshalAttributes(attrs)
	if err != nil {
		return fmt.Errorf("nested %d: %w", a.t, err)
	}

	// prevent copies of attributes buffer leaking outside the call.
	attrs.buf = nil

	return nil
}

// UnmarshalBytes is called to read the raw contents of an attribute into a
// type implementing [encoding.BinaryUnmarshaler]. It is the responsibility of
// that implementation to handle decoding from the host byteorder.
//
// WARNING: the bytes given to [encoding.BinaryUnmarshaler] will be a slice of
// of the message buffer, and is only valid for the duration of the message
// being unmarshaled. Consider using [Attribute.Bytes] or [Attribute.Copy].
func (a Attribute) UnmarshalBytes(dst encoding.BinaryUnmarshaler) error {
	if dst == nil {
		return fmt.Errorf("attribute %d: BinaryUnmarshaler is nil", a.t)
	}

	err := dst.UnmarshalBinary(a.buf)
	if err != nil {
		return fmt.Errorf("attribute %d: %w", a.t, err)
	}

	return nil
}

func (a *Attribute) unmarshal(b []byte) error {
	if len(b) < attrHdrLen {
		return fmt.Errorf("attribute: expected at least %d bytes, got %d", attrHdrLen, len(b))
	}

	length := int(binary.NativeEndian.Uint16(b))
	a.t = binary.NativeEndian.Uint16(b[2:])

	if length < attrHdrLen {
		// length does not include the header itself.
		return fmt.Errorf("attribute %d: invalid length, expected at least %d bytes, got %d", a.t, attrHdrLen, length)
	}

	if len(b) < length {
		// not enough bytes for attributed.
		return fmt.Errorf("attribute %d: needed %d bytes, got %d", a.t, length, len(b))
	}

	a.buf = b[attrHdrLen:length]
	return nil
}

// AttributeUnmarshaler is used to unmarshal the attributes or nested
// attributes contained within a Netlink message.
type AttributeUnmarshaler interface {
	UnmarshalAttributes(*AttributeReader) error
}

// AttributeUnmarshalerFunc adapts a function implementing the same signature
// as [AttributeUnmarshaler.UnmarshalAttributes] into an
// [AttributeUnmarshaler].
type AttributeUnmarshalerFunc func(*AttributeReader) error

// UnmarshalAttributes calls fn(attrs).
func (fn AttributeUnmarshalerFunc) UnmarshalAttributes(attrs *AttributeReader) error {
	return fn(attrs)
}

// AttributeReader is used to iterate through the Length-Tag-Value (LTV)
// attributes received as part of a Netlink message, using the host native
// byteorder.
//
// Intermediate family-specific headers should be read before progressing to
// attribute handling.
type AttributeReader struct {
	i   int
	buf []byte
	err error
}

// NewAttributeReader initializes a new [AttributeReader] from bytes containing
// the attributes to read, and optional intermediate family-specific headers
// for [AttributeReader.Read].
func NewAttributeReader(buf []byte) *AttributeReader {
	return &AttributeReader{buf: buf}
}

// Err returns the last error encountered while reading attributes, if any.
func (ar *AttributeReader) Err() error {
	return ar.err
}

// Each is an [iter.Seq] iterator, that will yield for each [Attribute]
// contained within the [AttributeReader], until finished or an error occurs.
//
// If an error occurs, it will be returned by [AttributeReader.Err].
func (ar *AttributeReader) Each(yield func(Attribute) bool) {
	if ar.err != nil {
		// attributes reader invalidated by previous error.
		return
	}

	// work on our own slice of the buffer to support multiple iterations.
	buf := ar.buf[ar.i:]

	for len(buf) > attrHdrLen {
		attr := Attribute{}
		err := attr.unmarshal(buf)
		if err != nil {
			ar.err = err
			break
		}

		if !yield(attr) {
			break
		}

		// progress the buffer for the next iteration, or end iteration.
		buf = buf[Align(attr.Length()):]
	}
}

// Length returns the total number of bytes contained within the
// [AttributeReader] that haven't been read.
func (ar *AttributeReader) Length() int {
	return len(ar.buf)
}

// Read arbitrary bytes from the Netlink attributes body. This is used to
// intercept intermediate family-specific headers before moving on to the
// attributes that follow.
//
// The bytes read will not automatically be aligned to 4 bytes, consider using
// [Align] to read the correct number of bytes.
func (ar *AttributeReader) Read(b []byte) (int, error) {
	if ar.i >= len(ar.buf) {
		return 0, io.EOF
	}

	n := copy(b, ar.buf[ar.i:])
	ar.i += n

	return n, nil
}

// Unmarshal is called on an [AttributeReader] to unmarshal it's attributes to
// a type implementing the [AttributeUnmarshaler] interface.
//
// If it contains any intermediate family-specific headers, these must be read
// first by calling [AttributeReader.Read].
func (ar *AttributeReader) Unmarshal(au AttributeUnmarshaler) error {
	if au == nil {
		return fmt.Errorf("attributes: AttributeUnmarshaler is nil")
	}

	err := au.UnmarshalAttributes(ar)
	if err != nil {
		return fmt.Errorf("attributes: %w", err)
	}

	return nil
}

// AttributeMarshaler is used to marshal Netlink attributes or nested
// attributes for a Netlink message.
//
// The total length of nested attributes cannot exceed a [math.Uint16].
type AttributeMarshaler interface {
	MarshalAttributes(*AttributeWriter) error
}

// AttributeMarshalerFunc adapts a function implementing the same signature as
// [AttributeMarshaler.MarshalAttributes] into an [AttributeMarshaler].
type AttributeMarshalerFunc func(*AttributeWriter) error

// MarshalAttributes calls fn(attrs).
func (fn AttributeMarshalerFunc) MarshalAttributes(attrs *AttributeWriter) error {
	return fn(attrs)
}

// AttributeWriter is used to build the Length-Tag-Value (LTV) attributes to be
// contained within a Netlink message, using the host native byteorder.
//
// Intermediate family-specific headers should be written before progressing to
// writing attributes.
type AttributeWriter struct {
	buf []byte
}

// NewAttributeWriter initializes a new [AttributeWriter] for marshaling a
// Netlink messages attributes, and optional intermediate family-specific
// headers.
//
// It may be given a buffer which already contains data, and/or a pre-allocated
// capacity which will be appended to.
func NewAttributeWriter(buf []byte) *AttributeWriter {
	return &AttributeWriter{buf: buf}
}

// AddBytes appends bytes attribute to the [AttributeWriter].
//
// An error is returned if the length of the bytes, plus the length of the
// attribute header, exceeds a [math.Uint16].
func (aw *AttributeWriter) AddBytes(attrType uint16, b []byte) error {
	length := attrHdrLen + len(b)
	if length > math.MaxUint16 {
		return fmt.Errorf("attribute: bytes exceeds uint16, got %d", length)
	}

	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, uint16(length))
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)

	aw.buf = append(aw.buf, b...)
	aw.buf = Pad(aw.buf)

	return nil
}

// AddInt8 appends an int8 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddInt8(attrType uint16, v int8) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 5)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = append(aw.buf, uint8(v), 0x00, 0x00, 0x00) //nolint
}

// AddInt16 appends an int16 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddInt16(attrType uint16, v int16) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 6)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, uint16(v)) //nolint
	aw.buf = append(aw.buf, 0x00, 0x00)
}

// AddInt32 appends an int32 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddInt32(attrType uint16, v int32) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 8)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint32(aw.buf, uint32(v)) //nolint
}

// AddInt64 appends an int64 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddInt64(attrType uint16, v int64) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 12)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint64(aw.buf, uint64(v)) //nolint
}

// AddString appends a null-terminated string attribute to the
// [AttributeWriter].
//
// An error is returned if the length of the string, plus the length of the
// attribute header and null-terminator, exceeds a [math.Uint16].
func (aw *AttributeWriter) AddString(attrType uint16, s string) error {
	length := attrHdrLen + len(s) + 1
	if length > math.MaxUint16 {
		return fmt.Errorf("attribute: string exceeds uint16, got %d", length)
	}

	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, uint16(length))
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)

	aw.buf = append(aw.buf, s...)
	aw.buf = append(aw.buf, 0x00)
	aw.buf = Pad(aw.buf)

	return nil
}

// AddUint8 appends a uint8 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddUint8(attrType uint16, v uint8) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 5)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = append(aw.buf, v, 0x00, 0x00, 0x00)
}

// AddUint16 appends a uint16 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddUint16(attrType uint16, v uint16) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 6)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, v)
	aw.buf = append(aw.buf, 0x00, 0x00)
}

// AddUint32 appends a uint32 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddUint32(attrType uint16, v uint32) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 8)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint32(aw.buf, v)
}

// AddUint64 appends a uint64 attribute to the [AttributeWriter].
func (aw *AttributeWriter) AddUint64(attrType uint16, v uint64) {
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, 12)
	aw.buf = binary.NativeEndian.AppendUint16(aw.buf, attrType)
	aw.buf = binary.NativeEndian.AppendUint64(aw.buf, v)
}

// Length returns the total number of bytes that have accumulated within the
// [AttributeWriter].
func (aw *AttributeWriter) Length() int {
	return len(aw.buf)
}

// Write arbitrary bytes to the Netlink attributes body. This is used to
// prepend intermediate family-specific headers before the attributes are
// marshaled.
//
// The bytes written will nto automatically aligned to 4 bytes, consider using
// [Pad] to write the correct number of bytes.
func (aw *AttributeWriter) Write(b []byte) (int, error) {
	aw.buf = append(aw.buf, b...)
	return len(b), nil
}
