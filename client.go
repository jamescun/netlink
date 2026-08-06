// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"fmt"
	"sync"
)

// Marshaler is implemented by types that can marshal themselves into a Netlink
// message, including type, flags, and attributes.
type Marshaler interface {
	MarshalNetlink(*Message) error
}

// MarshalerFunc adapts a function implementing the same signature as
// [Marshaler.MarshalNetlink] into a [Marshaler].
type MarshalerFunc func(*Message) error

// MarshalNetlink calls fn(msg).
func (fn MarshalerFunc) MarshalNetlink(msg *Message) error {
	return fn(msg)
}

// Unmarshaler is implemented by types that can unmarshal themselves from a
// Netlink message.
//
// If a response contains multiple messages, the unmarshaler will be called
// multiple times, until the whole response is consumed.
type Unmarshaler interface {
	UnmarshalNetlink(*Message) error
}

// UnmarshalerFunc adapts a function implementing the same signature as
// [Unmarshaler.UnmarshalNetlink] into an [Unmarshaler].
type UnmarshalerFunc func(*Message) error

// UnmarshalNetlink calls fn(msg).
func (fn UnmarshalerFunc) UnmarshalNetlink(msg *Message) error {
	return fn(msg)
}

// Client is a wrapper around [Conn] for exchanging client requests and
// responses with the unix socket, as well as managing the temporary buffers
// used between requests and incrementing sequence numbers.
//
// It is safe for concurrent use, the client will only ever have one request
// and response exchange in-flight at once.
type Client interface {
	// Close the client, underlying socket, and prevent further client
	// exchanges.
	Close() error

	// Family returns the Netlink [Family] this client was opened with.
	Family() Family

	// Seq returns the last sequence number used by the client.
	Seq() uint32

	// Do executes a Netlink request, marshaling the request and unmarshaling
	// the response(s) to dst.
	//
	// If the response contains multiple messages, dst should expect to handle
	// multiple messages, until it receives the [DONE] message type.
	//
	// If an [ERROR] type message is received, it will be unmarshaled and
	// returned as the [Error] type.
	//
	// The length, sequence number and port id will automatically be set.
	Do(Marshaler, Unmarshaler) error
}

type client struct {
	conn Conn
	seq  uint32
	pool sync.Pool
	mu   sync.Mutex
}

// Connect establishes a Netlink socket connection for the given [Family], and
// creates a [Client] for exchanging messages with request-response semantics.
func Connect(family Family) (Client, error) {
	conn, err := Dial(family)
	if err != nil {
		return nil, err
	}

	return NewClient(conn), nil
}

// NewClient initializes a [Client] around an already established [Conn], for
// exchanging messages with request-response semantics.
//
// This may be useful for testing and development, such as using a [Dumper]
// conn implementation, or otherwise intercept used to intercept reads and
// writes.
func NewClient(conn Conn) Client {
	return &client{
		conn: conn,
		pool: sync.Pool{
			New: func() any {
				// documentation suggests sizing this at least the page size,
				// or at least 32KB for dumping.
				//
				// See: https://www.kernel.org/doc/html/next/userspace-api/netlink/intro.html#buffer-sizing
				b := make([]byte, 32*1024)
				return &b
			},
		},
	}
}

func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.Close()
}

func (c *client) Family() Family {
	return c.conn.Family()
}

func (c *client) Seq() uint32 {
	return c.seq
}

func (c *client) Do(src Marshaler, dst Unmarshaler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO(jc): use shared buffer as well, can be the same buffer as below,
	// this is fine for now, as request messages tend to be small.
	msg := &Message{}
	err := src.MarshalNetlink(msg)
	if err != nil {
		return fmt.Errorf("client: marshal: %w", err)
	}

	c.seq++

	msg.Header.Seq = c.seq
	msg.Header.Pid = c.conn.Pid()

	buf, err := msg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("client: marshal: %w", err)
	}

	_, err = c.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("client: write: %w", err)
	}

	bufp, _ := c.pool.Get().(*[]byte)
	defer c.pool.Put(bufp)

	buf = *bufp

	n, err := c.conn.Read(buf)
	if err != nil {
		return fmt.Errorf("client: read: %w", err)
	}

	buf = buf[:n]

	for len(buf) > hdrLen {
		msg := &Message{}
		err := msg.UnmarshalBinary(buf)
		if err != nil {
			return fmt.Errorf("client: unmarshal: %w", err)
		}

		if msg.Type == ERROR {
			nlerr := &Error{}
			err = msg.Unmarshal(nlerr)
			if err == nil {
				err = nlerr
			}

			return err
		}

		err = dst.UnmarshalNetlink(msg)
		if err != nil {
			return fmt.Errorf("client: unmarshal: %w", err)
		}

		buf = buf[Align(msg.Length):]
	}

	return nil
}
