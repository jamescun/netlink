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
	"go.jamescun.com/netlink/genetlink"

	"golang.org/x/sys/unix"
)

func TestClient(t *testing.T) {
	// This tests against the generic netlink controller, as it is the only one
	// that is guaranteed to exist and it's family number is static.

	client, err := netlink.NewClient(netlink.GENERIC)
	if err != nil {
		t.Fatal("could not connect to netlink:", err)
	}
	defer client.Close()

	if client.Family() != netlink.GENERIC {
		t.Errorf("expected family GENERIC, got %s", client.Family())
	}

	err = client.Get(
		unix.GENL_ID_CTRL,
		netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
			err := msg.MarshalBytes(&genetlink.Header{
				Cmd:     unix.CTRL_CMD_GETFAMILY,
				Version: 2,
			})
			if err != nil {
				return err
			}

			err = msg.Marshal(netlink.AttributeMarshalerFunc(func(attrs *netlink.AttributeEncoder) error {
				attrs.Uint16(unix.CTRL_ATTR_FAMILY_ID, unix.GENL_ID_CTRL)
				return nil
			}))
			if err != nil {
				return err
			}

			return nil
		}),
		netlink.UnmarshalerFunc(func(msg netlink.MessageDecoder) error {
			if msg.Header().Type != 16 {
				t.Errorf("expected type 16 got %d", msg.Header().Type)
			}

			if msg.Header().Flags != 0 {
				t.Errorf("expected flags 0, got %d", msg.Header().Flags)
			}

			hdr := &genetlink.Header{}
			err := msg.UnmarshalBytes(4, hdr)
			if err != nil {
				return err
			}

			if hdr.Cmd != 1 {
				t.Errorf("expected genetlink cmd 1, got %d", hdr.Cmd)
			}

			if hdr.Version != 2 {
				t.Errorf("expected genetlink version 2, got %d", hdr.Version)
			}

			err = msg.Unmarshal(netlink.AttributeUnmarshalerFunc(func(attrs *netlink.AttributeDecoder) error {
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
				return err
			}

			return nil
		}),
	)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
