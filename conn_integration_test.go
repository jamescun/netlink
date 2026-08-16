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
)

func TestDial(t *testing.T) {
	// This tests against the generic netlink controller, as it is the only one
	// that is guaranteed to exist and it's family number is static.

	conn, err := netlink.Dial(netlink.GENERIC)
	if err != nil {
		t.Fatal("could not connect to netlink:", err)
	}
	defer conn.Close()

	if conn.Family() != netlink.GENERIC {
		t.Errorf("expected family GENERIC, got %s", conn.Family())
	}

	if conn.Pid() == 0 {
		t.Error("expected non-zero pid")
	}

	out := []byte{
		0x1c, 0x00, 0x00, 0x00, // length.
		0x10, 0x00, 0x05, 0x00, // GENL_ID_CTRL, REQUEST|ACK.
		0x00, 0x00, 0x00, 0x00, // sequent number.
		0x00, 0x00, 0x00, 0x00, // port id.
		0x03, 0x01, 0x00, 0x00, // genetlink CMD_GET_FAMILY version 1.

		// CTRL_ATTR_FAMILY_ID 0x10
		0x06, 0x00, 0x01, 0x00,
		0x10, 0x00, 0x00, 0x00,
	}

	n, err := conn.Write(out)
	if err != nil {
		t.Fatal("unexpected write error:", err)
	}

	if n != 28 {
		t.Errorf("expected 28 byte written, got %d", n)
	}

	in := make([]byte, 256)

	n, err = conn.Read(in)
	if err != nil {
		t.Fatal("unexpected read error:", err)
	}

	if n != 136 {
		t.Fatalf("expected 136 bytes read, got %d", n)
	}
}
