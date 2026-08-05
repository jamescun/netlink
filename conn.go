// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

// Conn is implemented to exchange message with the Linux Kernel using Netlink,
// either as a client of another subsystem, or as a server through genetlink.
//
// It will be automatically assigned a pid. It does not manage sequence
// numbers.
//
// The extended acknowledgement [NETLINK_EXT_ACK] will be enabled
// automatically, although this may be ignored depending on family.
type Conn interface {
	// Close the conn, no more messages can be read or written.
	Close() error

	// Family returns the Netlink [Family] this conn was opened with.
	Family() Family

	// Pid returns the port id (sometimes called the process id) that was
	// automatically assigned to this conn.
	Pid() uint32

	// Read one or more Netlink messages into bytes.
	//
	// Netlink communicates using datagrams, if the response to a request does
	// not fit into bytes, and error will be returned.
	Read(b []byte) (int, error)

	// Write one or more Netlink messages from bytes.
	//
	// Netlink communicates using datagrams, each request message must be
	// contained within the bytes.
	Write(b []byte) (int, error)
}

// Dial opens a connection to a Netlink socket with the specified [Family].
func Dial(family Family) (Conn, error) {
	return dial(family)
}
