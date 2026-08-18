// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sys/unix"
)

func TestIfInfoMsg(t *testing.T) {
	t.Run("MarshalBinary", func(t *testing.T) {
		tests := []struct {
			desc     string
			msg      IfInfoMsg
			expected []byte
		}{
			{"Zero", IfInfoMsg{}, []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}},
			{"InvalidIndex", IfInfoMsg{Index: math.MaxUint32}, nil},
			{"OK", IfInfoMsg{
				Family:  unix.AF_INET,
				Type:    EtherType(unix.ARPHRD_LOOPBACK),
				Index:   1234,
				Flags:   UP | NO_ARP,
				Changed: UP | NO_ARP,
			}, []byte{
				0x02, 0x00, 0x04, 0x03, 0xd2, 0x04, 0x00, 0x00,
				0x81, 0x00, 0x00, 0x00, 0x81, 0x00, 0x00, 0x00,
			}},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target, err := test.msg.MarshalBinary()

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

	t.Run("UnmarshalBinary", func(t *testing.T) {
		tests := []struct {
			desc     string
			bytes    []byte
			expected *IfInfoMsg
		}{
			{"Nil", nil, nil},
			{"Empty", []byte{}, nil},
			{"Zero", []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}, &IfInfoMsg{}},
			{"OK", []byte{
				0x02, 0x00, 0x04, 0x03, 0xd2, 0x04, 0x00, 0x00,
				0x81, 0x00, 0x00, 0x00, 0x81, 0x00, 0x00, 0x00,
			}, &IfInfoMsg{
				Family:  unix.AF_INET,
				Type:    EtherType(unix.ARPHRD_LOOPBACK),
				Index:   1234,
				Flags:   UP | NO_ARP,
				Changed: UP | NO_ARP,
			}},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target := &IfInfoMsg{}

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
						t.Error(test.expected, target)
					}
				}
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			desc     string
			msg      IfInfoMsg
			expected string
		}{
			{"Zero", IfInfoMsg{}, ";; ->>IFINFOMSG<<- family: 0, type: NETROM, index: 0\n;; flags: NONE, changed: NONE\n"},
			{"OK", IfInfoMsg{
				Family:  unix.AF_INET,
				Type:    EtherType(unix.ARPHRD_LOOPBACK),
				Index:   1234,
				Flags:   UP | NO_ARP,
				Changed: UP | NO_ARP,
			}, ";; ->>IFINFOMSG<<- family: 2, type: LOOPBACK, index: 1234\n;; flags: UP NO_ARP, changed: UP NO_ARP\n"},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				target := test.msg.String()

				if !cmp.Equal(test.expected, target) {
					t.Error(cmp.Diff(test.expected, target))
				}
			})
		}
	})
}
