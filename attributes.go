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
	"math"
	"net"
	"net/netip"
)

// attrHeaderLen is the length of the attribute header in bytes.
const attrHeaderLen = 4

// Constants for attribute types.
const (
	NESTED        uint16 = (1 << 15)
	NET_BYTEORDER uint16 = (1 << 14)
	TYPE_MASK     uint16 = ^(NESTED | NET_BYTEORDER)
)

// AttributeDecoder is used to decode a Length-Tag-Value (LTV) attribute from
// bytes inside a Netlink message, containing a number of convenience methods
// for nested attributes, common data types and interfaces.
type AttributeDecoder struct {
	attrType uint16
	buf      []byte
	err      error
}

// NewAttributeDecoder initializes a new [AttributeDecoder] from host byteorder
// bytes, including the attribute header.
//
// It will ignore any additional bytes it is given.
func NewAttributeDecoder(buf []byte) (*AttributeDecoder, error) {
	if len(buf) < attrHeaderLen {
		return nil, fmt.Errorf("needed at least %d bytes, got %d", attrHeaderLen, len(buf))
	}

	length := int(binary.NativeEndian.Uint16(buf))

	if length < attrHeaderLen {
		return nil, fmt.Errorf("invalid length, must be at least %d bytes, got %d", attrHeaderLen, length)
	} else if len(buf) < length {
		return nil, fmt.Errorf("needed %d bytes, got %d", length, len(buf))
	}

	return &AttributeDecoder{
		attrType: binary.NativeEndian.Uint16(buf[2:]),
		buf:      buf[attrHeaderLen:length],
	}, nil
}

// Length returns the unaligned length of the attribute, not including the
// header.
func (ad *AttributeDecoder) Length() int {
	return len(ad.buf)
}

// Type returns the type of the attribute from the attribute header.
//
// If the type has either the NESTED or NET_BYTEORDER type flags set, these
// will not be returned.
func (ad *AttributeDecoder) Type() uint16 {
	return ad.attrType & TYPE_MASK
}

// IsNested returns true if the attribute type has the [NESTED] flag set.
func (ad *AttributeDecoder) IsNested() bool {
	return ad.attrType&NESTED != 0
}

// IsNetByteOrder returns true if the attribute type has the [NET_BYTEORDER]
// flag set.
func (ad *AttributeDecoder) IsNetByteOrder() bool {
	return ad.attrType&NET_BYTEORDER != 0
}

// Err returns the last error encountered while iterating through a nested
// attribute, if any.
func (ad *AttributeDecoder) Err() error {
	return ad.err
}

// Addr unmarshals the attribute as a [netip.Addr] IP address.
func (ad *AttributeDecoder) Addr() (netip.Addr, bool) {
	return netip.AddrFromSlice(ad.buf)
}

// Bool interprets any integer with a value greater than zero as a boolean
// true, including flags with no value.
func (ad *AttributeDecoder) Bool() bool {
	if len(ad.buf) == 0 {
		// flags are true if present but have no body.
		return true
	}

	for _, b := range ad.buf {
		if b > 0 {
			return true
		}
	}

	return false
}

// Bytes returns the contents of the attribute as bytes, excluding the
// attribute header or any alignment bytes.
func (ad *AttributeDecoder) Bytes() []byte {
	b := make([]byte, len(ad.buf))
	copy(b, ad.buf)
	return b
}

// Copy the contents of the attributes bytes into dst, returning how many
// bytes were copied.
//
// Multiple calls to Copy will yield the same bytes.
func (ad *AttributeDecoder) Copy(dst []byte) int {
	return copy(dst, ad.buf)
}

// HardwareAddr unmarshals the attribute a [net.HardwareAddr] MAC address.
func (ad *AttributeDecoder) HardwareAddr() net.HardwareAddr {
	return net.HardwareAddr(ad.Bytes())
}

// Int8 unmarshal the attribute as an int8.
//
// If not an int8, zero is returned.
func (ad *AttributeDecoder) Int8() int8 {
	return int8(ad.Uint8()) //nolint
}

// Int16 unmarshal the attribute as an int16.
//
// If not an int16, zero is returned.
func (ad *AttributeDecoder) Int16() int16 {
	return int16(ad.Uint16()) //nolint
}

// Int32 unmarshal the attribute as an int32.
//
// If not an int32, zero is returned.
func (ad *AttributeDecoder) Int32() int32 {
	return int32(ad.Uint32()) //nolint
}

// Int64 unmarshal the attribute as an int64.
//
// If not an int64, zero is returned.
func (ad *AttributeDecoder) Int64() int64 {
	return int64(ad.Uint64()) //nolint
}

// Each is an [iter.Seq] iterator that will yield for each nested attribute or
// nested array contained within an attribute.
//
// If an error is encountered while iterating, yielding will stop and the error
// returned by [AttributeDecoder.Err].
func (ad *AttributeDecoder) Each(yield func(*AttributeDecoder) bool) {
	if ad.err != nil {
		// invalidated by previous error.
		return
	}

	// work on our own slice of the buffer, for multiple iterations.
	buf := ad.buf

	i := 0
	for len(buf) >= attrHeaderLen {
		attr, err := NewAttributeDecoder(buf)
		if err != nil {
			ad.err = fmt.Errorf("attribute %d: %w", i, err)
			return
		}

		if !yield(attr) {
			return
		}

		length := Align(attrHeaderLen + attr.Length())
		if length > len(buf) {
			ad.err = fmt.Errorf("attribute %d unaligned, expected %d, got %d", i, length, len(buf))
			return
		}

		i++
		buf = buf[length:]
	}

	if len(buf) != 0 {
		ad.err = fmt.Errorf("attribute had %d bytes unread", len(buf))
	}
}

// String returns the contents of the attribute as a string, not including the
// null-terminator.
//
// If not a string, the string will contain gibberish.
func (ad *AttributeDecoder) String() string {
	if len(ad.buf) == 0 {
		return ""
	}

	buf := ad.buf

	// trim null terminator, if it has one, before casting to a string.
	if buf[len(buf)-1] == 0x00 {
		buf = buf[:len(buf)-1]
	}

	return string(buf)
}

// Uint8 unmarshal the attribute as a uint8.
//
// If not a uint8, zero is returned.
func (ad *AttributeDecoder) Uint8() uint8 {
	if len(ad.buf) < 1 {
		return 0
	}

	return ad.buf[0]
}

// Uint16 unmarshal the attribute as a uint16.
//
// If not a uint16, zero is returned.
func (ad *AttributeDecoder) Uint16() uint16 {
	if len(ad.buf) < 2 {
		return 0
	}

	return binary.NativeEndian.Uint16(ad.buf)
}

// Uint32 unmarshal the attribute as a uint32.
//
// If not a uint32, zero is returned.
func (ad *AttributeDecoder) Uint32() uint32 {
	if len(ad.buf) < 4 {
		return 0
	}

	return binary.NativeEndian.Uint32(ad.buf)
}

// Uint64 unmarshal the attribute as a uint64.
//
// If not a uint64, zero is returned.
func (ad AttributeDecoder) Uint64() uint64 {
	if len(ad.buf) < 8 {
		return 0
	}

	return binary.NativeEndian.Uint64(ad.buf)
}

// Unmarshal is called to read nested attributes from this attribute using an
// [AttributeUnmarshaler].
//
// You may also want to consider [AttributeDecoder.Nested] to iterate through
// the the attributes, if [AttributeUnmarshaler] is not implemented.
func (ad *AttributeDecoder) Unmarshal(au AttributeUnmarshaler) error {
	if au == nil {
		return fmt.Errorf("AttributeUnmarshaler is nil")
	}

	attrs := &AttributeDecoder{attrType: ad.attrType, buf: ad.buf}

	err := au.UnmarshalAttributes(attrs)
	if err != nil {
		return err
	}

	return attrs.Err()
}

// UnmarshalBytes is called to read the raw contents of an attribute into a
// type implementing [encoding.BinaryUnmarshaler]. It is the responsibility of
// that implementation to handle decoding from the host byteorder.
func (ad *AttributeDecoder) UnmarshalBytes(dst encoding.BinaryUnmarshaler) error {
	if dst == nil {
		return fmt.Errorf("encoding.BinaryUnmarshaler is nil")
	}

	err := dst.UnmarshalBinary(ad.buf)
	if err != nil {
		return err
	}

	return nil
}

// AttributeEncoder is used to encode a Length-Tag-Value (LTV) attribute to
// bytes for a Netlink message, containing a number of convenience methods
// for nested attributes, common data types and interfaces.
type AttributeEncoder struct {
	buf []byte
}

// Addr marshals an attribute containing a [netip.Addr].
func (ae *AttributeEncoder) Addr(attrType uint16, addr netip.Addr) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, uint16(attrHeaderLen+(addr.BitLen()/8))) //nolint
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)
	ae.buf, _ = addr.AppendBinary(ae.buf)
}

// Bytes marshals an attribute containing arbitrary bytes.
//
// The length of bytes cannot exceed a uint16, including the attribute header.
func (ae *AttributeEncoder) Bytes(attrType uint16, b []byte) error {
	length := len(b) + attrHeaderLen
	if length < 0 || length > math.MaxUint16 {
		return fmt.Errorf("bytes length exceeds uint16, got %d", length)
	}

	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, uint16(length))
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = append(ae.buf, b...)
	ae.buf = Pad(ae.buf)

	return nil
}

// HardwareAddr marshals an attribute containing a [net.HardwareAddr].
func (ae *AttributeEncoder) HardwareAddr(attrType uint16, addr net.HardwareAddr) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, uint16(len(addr)+attrHeaderLen)) //nolint
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = append(ae.buf, addr...)
	ae.buf = Pad(ae.buf)
}

// Int8 marshals an attribute containing an int8.
func (ae *AttributeEncoder) Int8(attrType uint16, v int8) {
	ae.Uint8(attrType, uint8(v)) //nolint
}

// Int16 marshals an attribute containing an int16.
func (ae *AttributeEncoder) Int16(attrType uint16, v int16) {
	ae.Uint16(attrType, uint16(v)) //nolint
}

// Int32 marshals an attribute containing an int32.
func (ae *AttributeEncoder) Int32(attrType uint16, v int32) {
	ae.Uint32(attrType, uint32(v)) //nolint
}

// Int64 marshals an attribute containing an int64.
func (ae *AttributeEncoder) Int64(attrType uint16, v int64) {
	ae.Uint64(attrType, uint64(v)) //nolint
}

// Marshal marshals a nested attribute from a type implementing
// [AttributeMarshaler].
//
// The length of the nested attributes cannot exceed a uint16, including the
// attribute header.
func (ae *AttributeEncoder) Marshal(attrType uint16, src AttributeMarshaler) error {
	if src == nil {
		return fmt.Errorf("AttributeMarshaler is nil")
	}

	// get initial length for length calculation after marshaling.
	start := len(ae.buf)

	attrs := &AttributeEncoder{buf: ae.buf}

	// append empty length to set after marshaling.
	attrs.buf = append(attrs.buf, 0x00, 0x00)
	attrs.buf = binary.NativeEndian.AppendUint16(attrs.buf, attrType)

	err := src.MarshalAttributes(attrs)
	if err != nil {
		return err
	}

	// calculate new length from initial length.
	length := len(attrs.buf) - start

	if length < 0 || length > math.MaxUint16 {
		return fmt.Errorf("nested length exceeds uint16, got %d", length)
	}

	binary.NativeEndian.PutUint16(attrs.buf[start:], uint16(length))

	ae.buf = attrs.buf

	return nil
}

// MarshalBytes marshals an attribute from arbitrary bytes from a type
// implementing [encoding.BinaryMarshaler]. It is the responsibility of
// that implementation to handle encoding to the host byteorder.
//
// The length of bytes cannot exceed a uint16, including the attribute header.
func (ae *AttributeEncoder) MarshalBytes(attrType uint16, src encoding.BinaryMarshaler) error {
	if src == nil {
		return fmt.Errorf("encoding.BinaryMarshaler is nil")
	}

	// get initial length for length calculation after marshaling.
	start := len(ae.buf)

	// append empty length to set after marshaling.
	buf := append(ae.buf, 0x00, 0x00) //nolint
	buf = binary.NativeEndian.AppendUint16(buf, attrType)

	if ba, ok := src.(encoding.BinaryAppender); ok {
		var err error
		buf, err = ba.AppendBinary(buf)
		if err != nil {
			return err
		}
	} else {
		b, err := src.MarshalBinary()
		if err != nil {
			return err
		}

		buf = append(buf, b...)
	}

	// calculate new length from initial length.
	length := len(buf) - start

	if length < 0 || length > math.MaxUint16 {
		return fmt.Errorf("bytes length exceeds uint16, got %d", length)
	}

	binary.NativeEndian.PutUint16(buf[start:], uint16(length))

	ae.buf = Pad(buf)

	return nil
}

// String marshals an attribute containing a null-terminated string.
//
// The length of string cannot exceed a uint16, including the attribute header
// and null-terminator.
func (ae *AttributeEncoder) String(attrType uint16, s string) error {
	length := len(s) + attrHeaderLen + 1
	if length < 0 || length > math.MaxUint16 {
		return fmt.Errorf("string length exceeds uint16, got %d", length)
	}

	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, uint16(length))
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = append(ae.buf, s...)
	ae.buf = append(ae.buf, 0x00)
	ae.buf = Pad(ae.buf)

	return nil
}

// Uint8 marshals an attribute containing a uint8.
func (ae *AttributeEncoder) Uint8(attrType uint16, v uint8) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrHeaderLen+1)
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = append(ae.buf, v, 0x00, 0x00, 0x00)
}

// Uint16 marshals an attribute containing a uint16.
func (ae *AttributeEncoder) Uint16(attrType uint16, v uint16) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrHeaderLen+2)
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, v)
	ae.buf = append(ae.buf, 0x00, 0x00)
}

// Uint32 marshals an attribute containing a uint32.
func (ae *AttributeEncoder) Uint32(attrType uint16, v uint32) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrHeaderLen+4)
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = binary.NativeEndian.AppendUint32(ae.buf, v)
}

// Uint64 marshals an attribute containing a uint64.
func (ae *AttributeEncoder) Uint64(attrType uint16, v uint64) {
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrHeaderLen+8)
	ae.buf = binary.NativeEndian.AppendUint16(ae.buf, attrType)

	ae.buf = binary.NativeEndian.AppendUint64(ae.buf, v)
}

// AttributeMarshaler is used to marshal attributes or nested attributes to a
// Netlink message.
type AttributeMarshaler interface {
	MarshalAttributes(*AttributeEncoder) error
}

// AttributeMarshalerFunc adapts a function implementing the same signature
// as [AttributeMarshaler.MarshalAttributes] into an [AttributeMarshaler].
type AttributeMarshalerFunc func(*AttributeEncoder) error

// MarshalAttributes calls fn(attrs).
func (fn AttributeMarshalerFunc) MarshalAttributes(attrs *AttributeEncoder) error {
	return fn(attrs)
}

// AttributeUnmarshaler is used to unmarshal the attributes or nested
// attributes contained within a Netlink message.
type AttributeUnmarshaler interface {
	UnmarshalAttributes(*AttributeDecoder) error
}

// AttributeUnmarshalerFunc adapts a function implementing the same signature
// as [AttributeUnmarshaler.UnmarshalAttributes] into an
// [AttributeUnmarshaler].
type AttributeUnmarshalerFunc func(*AttributeDecoder) error

// UnmarshalAttributes calls fn(attrs).
func (fn AttributeUnmarshalerFunc) UnmarshalAttributes(attrs *AttributeDecoder) error {
	return fn(attrs)
}
