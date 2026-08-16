// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"math"
	"net"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAttributeEncoder(t *testing.T) {
	t.Run("Addr/IPv4", func(t *testing.T) {
		expected := []byte{
			0x08, 0x00, 0x22, 0x00,
			0xC0, 0xA8, 0x01, 0x18,
		}

		ae := &AttributeEncoder{}

		ae.Addr(34, netip.AddrFrom4([4]byte{192, 168, 1, 24}))

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Bytes", func(t *testing.T) {
		expected := []byte{
			0x09, 0x00, 0x0A, 0x00,
			'h', 'e', 'l', 'l', 'o', 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		err := ae.Bytes(10, []byte("hello"))
		if err != nil {
			t.Fatal("unexpected error: %w", err)
		}

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("HardwareAddr", func(t *testing.T) {
		mac, _ := net.ParseMAC("f6:be:74:08:82:1c")

		expected := []byte{
			0x0A, 0x00, 0x0C, 0x00,
			0xF6, 0xBE, 0x74, 0x08, 0x82, 0x1C, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.HardwareAddr(12, mac)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int8/Min", func(t *testing.T) {
		expected := []byte{
			0x05, 0x00, 0x06, 0x00,
			0x80, 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Int8(6, math.MinInt8)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int8/Max", func(t *testing.T) {
		expected := []byte{
			0x05, 0x00, 0x06, 0x00,
			0x7F, 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Int8(6, math.MaxInt8)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int16/Min", func(t *testing.T) {
		expected := []byte{
			0x06, 0x00, 0x07, 0x00,
			0x00, 0x80, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Int16(7, math.MinInt16)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int16/Max", func(t *testing.T) {
		expected := []byte{
			0x06, 0x00, 0x07, 0x00,
			0xFF, 0x7F, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Int16(7, math.MaxInt16)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int32/Min", func(t *testing.T) {
		expected := []byte{
			0x08, 0x00, 0x08, 0x00,
			0x00, 0x00, 0x00, 0x80,
		}

		ae := &AttributeEncoder{}

		ae.Int32(8, math.MinInt32)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int32/Max", func(t *testing.T) {
		expected := []byte{
			0x08, 0x00, 0x08, 0x00,
			0xFF, 0xFF, 0xFF, 0x7F,
		}

		ae := &AttributeEncoder{}

		ae.Int32(8, math.MaxInt32)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int64/Min", func(t *testing.T) {
		expected := []byte{
			0x0C, 0x00, 0x09, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80,
		}

		ae := &AttributeEncoder{}

		ae.Int64(9, math.MinInt64)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Int64/Max", func(t *testing.T) {
		expected := []byte{
			0x0C, 0x00, 0x09, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F,
		}

		ae := &AttributeEncoder{}

		ae.Int64(9, math.MaxInt64)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		expected := []byte{
			0x10, 0x00, 0x0C, 0x80,

			0x0A, 0x00, 0x22, 0x00,
			'h', 'e', 'l', 'l', 'o', 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		err := ae.Marshal(12, AttributeMarshalerFunc(func(attrs *AttributeEncoder) error {
			return attrs.String(34, "hello")
		}))
		if err != nil {
			t.Fatal("unexpected error:", err)
		}

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("MarshalBytes", func(t *testing.T) {
		expected := []byte{
			0x08, 0x00, 0x22, 0x00,
			0xC0, 0xA8, 0x01, 0x18,
		}

		ae := &AttributeEncoder{}

		// netip.Addr also implements Append/MarshalBinary, can use this or
		// Addr, and should yield equivalent bytes.
		err := ae.MarshalBytes(34, netip.AddrFrom4([4]byte{192, 168, 1, 24}))
		if err != nil {
			t.Fatal("unexpected error:", err)
		}

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("String", func(t *testing.T) {
		expected := []byte{
			0x0A, 0x00, 0x0b, 0x00,
			'h', 'e', 'l', 'l', 'o', 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		err := ae.String(11, "hello")
		if err != nil {
			t.Fatal("unexpected error: %w", err)
		}

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Uint8", func(t *testing.T) {
		expected := []byte{
			0x05, 0x00, 0x02, 0x00,
			0xFF, 0x00, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Uint8(2, math.MaxUint8)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Uint16", func(t *testing.T) {
		expected := []byte{
			0x06, 0x00, 0x03, 0x00,
			0xFF, 0xFF, 0x00, 0x00,
		}

		ae := &AttributeEncoder{}

		ae.Uint16(3, math.MaxUint16)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Uint32", func(t *testing.T) {
		expected := []byte{
			0x08, 0x00, 0x04, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF,
		}

		ae := &AttributeEncoder{}

		ae.Uint32(4, math.MaxUint32)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		expected := []byte{
			0x0c, 0x00, 0x05, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}

		ae := &AttributeEncoder{}

		ae.Uint64(5, math.MaxUint64)

		if !cmp.Equal(expected, ae.buf) {
			t.Error(cmp.Diff(expected, ae.buf))
		}
	})
}
