// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAlign(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 4},
		{2, 4},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{9, 12},
		{-1, 0},
		{-2, 0},
	}

	for _, test := range tests {
		target := Align(test.n)

		if test.expected != target {
			t.Errorf("expected %d got %d", test.expected, target)
		}
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		b        []byte
		expected []byte
	}{
		{nil, nil},
		{[]byte{0x00}, []byte{0x00, 0x00, 0x00, 0x00}},
		{[]byte{0x00, 0x00}, []byte{0x00, 0x00, 0x00, 0x00}},
		{[]byte{0x00, 0x00, 0x00}, []byte{0x00, 0x00, 0x00, 0x00}},
		{[]byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00, 0x00, 0x00}},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{[]byte{'h', 'e', 'l', 'l', 'o'}, []byte{'h', 'e', 'l', 'l', 'o', 0x00, 0x00, 0x00}},
	}

	for _, test := range tests {
		target := Pad(test.b)

		if !cmp.Equal(test.expected, target) {
			t.Error(cmp.Diff(test.expected, target))
		}
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		flags    Flags
		expected string
	}{
		{0, "00"},
		{REQUEST, "00|REQUEST"},
		{REQUEST | ACK, "00|REQUEST|ACK"},
		{REQUEST | DUMP_INTR, "00|REQUEST|DUMP_INTR"},
		{REQUEST | ACK | DUMP_INTR, "00|REQUEST|ACK|DUMP_INTR"},
		{REQUEST | DUMP, "03|REQUEST"},
	}

	for _, test := range tests {
		target := test.flags.String()

		if test.expected != target {
			t.Errorf("expected:\n%s\ngot:\n%s", test.expected, target)
		}
	}
}

func TestHeader(t *testing.T) {
	t.Run("MarshalBinary", func(t *testing.T) {
		tests := []struct {
			desc     string
			header   *Header
			expected []byte
		}{
			{"invalid", &Header{Length: -1234}, nil},
			{"short", &Header{Length: 15}, nil},
			{"zero", &Header{}, []byte{
				0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}},
			{"dump", &Header{0, 0, REQUEST | DUMP, 1234, 5678}, []byte{
				0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x03,
				0xd2, 0x04, 0x00, 0x00, 0x2e, 0x16, 0x00, 0x00,
			}},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target, err := test.header.MarshalBinary()

				if test.expected == nil {
					if err == nil {
						t.Fatal("expected error")
					}
				} else {
					if err != nil {
						t.Fatal("unexpected error:", err)
					}

					if !bytes.Equal(test.expected, target) {
						t.Error(cmp.Diff(test.expected, target))
					}
				}
			})
		}
	})

	t.Run("UnmarshalBinary", func(t *testing.T) {
		tests := []struct {
			desc     string
			bytes    []byte
			expected *Header
		}{
			{"nil", nil, nil},
			{"empty", []byte{}, nil},
			{"zero", []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}, nil},
			{"header only", []byte{
				0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}, &Header{Length: 16}},
			{"dump", []byte{
				0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x03,
				0xd2, 0x04, 0x00, 0x00, 0x2e, 0x16, 0x00, 0x00,
			}, &Header{16, 0, REQUEST | DUMP, 1234, 5678}},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target := new(Header)
				err := target.UnmarshalBinary(test.bytes)

				if test.expected == nil {
					if err == nil {
						t.Fatal("expected error")
					}
				} else {
					if err != nil {
						t.Fatal("unexpected error:", err)
					}

					if !cmp.Equal(test.expected, target) {
						t.Error(cmp.Diff(test.expected, target))
					}
				}
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			desc     string
			header   *Header
			expected string
		}{
			{
				"dump",
				&Header{16, 0, REQUEST | DUMP, 1234, 5678},
				";; ->>NETLINK<<- type: 0, flags: 03|REQUEST\n;; length: 16, seq: 1234, pid: 5678\n",
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target := test.header.String()

				if target != test.expected {
					t.Error(cmp.Diff(test.expected, target))
				}
			})
		}
	})
}
