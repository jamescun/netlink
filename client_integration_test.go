// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

//go:build integration

package netlink_test

import (
	"testing"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

func TestClient(t *testing.T) {
	// This tests against the generic netlink controller, as it is the only one
	// that is guaranteed to exist and it's family number is static.

	client, err := netlink.Connect(netlink.GENERIC)
	if err != nil {
		t.Fatal("could not connect to netlink:", err)
	}
	defer client.Close()

	if client.Family() != netlink.GENERIC {
		t.Errorf("expected family GENERIC, got %s", client.Family())
	}

	err = client.Do(
		netlink.MarshalerFunc(func(msg *netlink.Message) error {
			msg.Type = 0x10 // GENL_ID_CTRL(16).
			msg.Flags = netlink.REQUEST | netlink.ACK

			msg.Write([]byte{
				0x03, 0x01, 0x00, 0x00, // genetlink CMD_GET_FAMILY version 1.

				// CTRL_ATTR_FAMILY_ID GENL_ID_CTRL(16).
				0x06, 0x00, 0x01, 0x00,
				0x10, 0x00, 0x00, 0x00,
			})

			return nil
		}),
		netlink.UnmarshalerFunc(func(msg *netlink.Message) error {
			if msg.Type != 16 {
				t.Errorf("expected type 16 got %d", msg.Type)
			}

			if msg.Flags != 0 {
				t.Errorf("expected flags 0, got %d", msg.Flags)
			}

			hdr := []byte{0x00, 0x00, 0x00, 0x00}
			n, err := msg.Read(hdr)
			if err != nil {
				t.Error("could not read genetlink header:", err)
			}
			if n != 4 {
				t.Errorf("expected 4 bytes for genetlink header, got %d", n)
			}

			if hdr[0] != 1 {
				t.Errorf("expected genetlink cmd 1, got %d", hdr[0])
			}

			if hdr[1] != 2 {
				t.Errorf("expected genetlink version 2, got %d", hdr[1])
			}

			err = msg.Unmarshal(netlink.AttributeUnmarshalerFunc(func(attrs *netlink.AttributeReader) error {
				for attr := range attrs.Each {
					switch attr.Type() {
					case unix.CTRL_ATTR_FAMILY_NAME:
						return nil
					}
				}

				t.Error("did not receive CTRL_ATTR_FAMILY_NAME")

				return nil
			}))
			if err != nil {
				t.Error("unexpected attribute error:", err)
			}

			return nil
		}),
	)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
