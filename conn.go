// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"errors"
	"time"
)

// ErrClosed is returned by [Conn] when attempting a read or write but the
// connection has been closed.
var ErrClosed = errors.New("conn closed")

// Conn is implemented to exchange message with the Linux Kernel using Netlink,
// either as a client of another subsystem, or as a server through genetlink.
//
// It will be automatically assigned a pid. It does not manage sequence
// numbers.
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

	// JoinGroup configures this connection to receive multicast notifications
	// from the specified group ID.
	//
	// It is recommended to use a connection dedicated to receiving
	// notifications, as the asynchronous nature of notifications may interfere
	// with other messages being receive at the same time.
	JoinGroup(group int) error

	// LeaveGroup stops this connection receiving multicast notifications from
	// a previously joined group.
	LeaveGroup(group int) error

	// SetDeadline sets the read and write deadlines associated with the
	// connection. It is equivalent to calling both SetReadDeadline and
	// SetWriteDeadline.
	SetDeadline(time.Time) error

	// SetReadDeadline sets the deadline for future Read calls and any
	// currently-blocked Read call.
	//
	// A zero value for t means Read will not time out.
	SetReadDeadline(time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls and any
	// currently-blocked Write call.
	//
	// Even if write times out, it may return n > 0, indicating that some of
	// the data was successfully written.
	//
	// A zero value for t means Write will not time out.
	SetWriteDeadline(time.Time) error
}

// Dial opens a connection to a Netlink socket with the specified [Family].
//
// The opened socket may optionally be configured by passing one-or-more
// [ConnOption].
func Dial(family Family, opts ...ConnOption) (Conn, error) {
	return dial(family, opts...)
}

// ConnOption is a function that allows for additional configuration of the
// file descriptor of the unix socket used to exchange Netlink messages.
type ConnOption func(fd int) error
