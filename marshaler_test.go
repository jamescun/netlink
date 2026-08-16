// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		desc     string
		target   Marshaler
		expected []byte
	}{
		{"Header", MessageHeader{
			Type:  16,
			Flags: REQUEST | DUMP,
		}, []byte{
			0x10, 0x00, 0x00, 0x00,
			0x10, 0x00, 0x01, 0x03,
			0xD2, 0x04, 0x00, 0x00,
			0x2E, 0x16, 0x00, 0x00,
		}},
		{"Header/Empty", MessageHeader{}, []byte{
			0x10, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xD2, 0x04, 0x00, 0x00,
			0x2E, 0x16, 0x00, 0x00,
		}},
		{"Generic", MarshalerFunc(func(msg MessageEncoder) error {
			msg.SetHeader(16, REQUEST|ACK)

			// generic netlink header.
			_, err := msg.Write([]byte{3, 2, 0, 0})
			if err != nil {
				return err
			}

			err = msg.Marshal(AttributeMarshalerFunc(func(attrs *AttributeEncoder) error {
				return attrs.String(2, "nlctrl")
			}))
			if err != nil {
				return err
			}

			return nil
		}), []byte{
			// message header.
			0x20, 0x00, 0x00, 0x00,
			0x10, 0x00, 0x05, 0x00,
			0xD2, 0x04, 0x00, 0x00,
			0x2E, 0x16, 0x00, 0x00,

			// generic netlink header.
			0x03, 0x02, 0x00, 0x00,

			// attributes.
			0x0B, 0x00, 0x02, 0x00,
			'n', 'l', 'c', 't', 'r', 'l', 0x00, 0x00,
		}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			target, err := Marshal(1234, 5678, test.target)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			if !cmp.Equal(test.expected, target) {
				t.Error(cmp.Diff(test.expected, target))
			}
		})
	}
}
