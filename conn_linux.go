// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"fmt"
	"net"
	"sync"

	"golang.org/x/sys/unix"
)

type conn struct {
	fd     int // -1 if closed.
	family Family
	pid    uint32
	mu     sync.Mutex
}

func dial(family Family) (Conn, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, int(family))
	if err != nil {
		return nil, fmt.Errorf("dial: socket: %w", err)
	}

	c := &conn{fd: fd, family: family}

	err = unix.Bind(c.fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
	})
	if err != nil {
		unix.Close(c.fd) //nolint
		return nil, fmt.Errorf("dial: bind: %w", err)
	}

	// get socket name to retrieve the port id.
	sa, err := unix.Getsockname(c.fd)
	if err != nil {
		unix.Close(c.fd) //nolint
		return nil, fmt.Errorf("dial: sockname: %w", err)
	}

	// enable extended acknowledgements, may not be implemented by the family.
	err = unix.SetsockoptInt(c.fd, unix.SOL_NETLINK, unix.NETLINK_EXT_ACK, 1)
	if err != nil {
		unix.Close(c.fd) //nolint
		return nil, fmt.Errorf("dial: setsockopt: %w", err)
	}

	if nl, ok := sa.(*unix.SockaddrNetlink); ok {
		c.pid = nl.Pid
	}

	return c, nil
}

func (c *conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fd == -1 {
		// already closed.
		return net.ErrClosed
	}

	err := unix.Close(c.fd)
	c.fd = -1
	return err
}

func (c *conn) Family() Family {
	return c.family
}

func (c *conn) Pid() uint32 {
	return c.pid
}

func (c *conn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fd == -1 {
		return 0, net.ErrClosed
	}

	n, _, flags, _, err := unix.Recvmsg(c.fd, b, nil, 0)
	if err != nil {
		return 0, fmt.Errorf("read: %w", err)
	} else if flags&unix.MSG_TRUNC != 0 {
		return 0, fmt.Errorf("read: truncated, exceeded buffer size %d", len(b))
	}

	return n, nil
}

func (c *conn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fd == -1 {
		return 0, net.ErrClosed
	}

	err := unix.Sendto(c.fd, b, 0, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
	})
	if err != nil {
		return 0, fmt.Errorf("write: %w", err)
	}

	return len(b), nil
}

func (c *conn) String() string {
	return fmt.Sprintf(
		"Conn(netlink, family: %s, fd: %d)",
		c.family, c.fd,
	)
}
