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

	// Do executes a Netlink request, marshaling the request from an
	// [Marshaler], and does not expect a response other than an ERROR
	// message type or an acknowledgement.
	//
	// The message will automatically have the message type configured as
	// msgType, and have the REQUEST, ACK and any additional flags set.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Do(msgType, flags uint16, src Marshaler) error

	// Dump executes a Netlink request, marshaling the request from a
	// [Marshaler], which may be nil, and unmarshaling the response(s) to an
	// [Unmarshaler].
	//
	// The message will automatically have the message type configured as
	// msgType, and have the REQUEST, DUMP and any additional flags set.
	//
	// If the response contains multiple messages, dst will be unmarshaled to
	// repeatedly until all messages have been consumed, not including the
	// DONE message type.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Dump(msgType, flags uint16, src Marshaler, dst Unmarshaler) error

	// Get executes a Netlink request, marshaling the request from a
	// [Marshaler], which may be nil, and unmarshaling the response to an
	// [Unmarshaler].
	//
	// The message will automatically have the message type configured as
	// msgType, and have the REQUEST, ACK and any additional flags set.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Get(msgType, flags uint16, src Marshaler, dst Unmarshaler) error
}

type client struct {
	conn Conn
	seq  uint32
	mu   sync.Mutex
}

// bufPool is a shared pool of buffers for marshaling and unmarshaling messages
// in [Client] and [Subscription].
//
// It is configured for 32KB, which should be big enough for large messages and
// dumping.
var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 32*1024)
		return &b
	},
}

// NewClient establishes a Netlink socket connection for the given [Family],
// and creates a [Client] for exchanging messages with request-response
// semantics.
//
// It may optionally be given [ConnOption] to configure the underlying socket.
func NewClient(family Family, opts ...ConnOption) (Client, error) {
	conn, err := Dial(family, opts...)
	if err != nil {
		return nil, err
	}

	return &client{
		conn: conn,
	}, nil
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

func (c *client) Do(msgType, flags uint16, src Marshaler) error {
	if src == nil {
		return fmt.Errorf("Marshaler is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.do(MarshalerFunc(func(msg MessageEncoder) error {
		msg.SetHeader(msgType, REQUEST|ACK|flags)
		return src.MarshalNetlink(msg)
	}), nil)
}

func (c *client) Dump(msgType, flags uint16, src Marshaler, dst Unmarshaler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if src == nil {
		return c.do(MessageHeader{
			Type:  msgType,
			Flags: REQUEST | DUMP | flags,
		}, dst)
	}

	return c.do(MarshalerFunc(func(msg MessageEncoder) error {
		msg.SetHeader(msgType, REQUEST|DUMP|flags)
		return src.MarshalNetlink(msg)
	}), dst)
}

func (c *client) Get(msgType, flags uint16, src Marshaler, dst Unmarshaler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if src == nil {
		return c.do(MessageHeader{
			Type:  msgType,
			Flags: REQUEST | ACK | flags,
		}, dst)
	}

	return c.do(MarshalerFunc(func(msg MessageEncoder) error {
		msg.SetHeader(msgType, REQUEST|ACK|flags)
		return src.MarshalNetlink(msg)
	}), dst)
}

// client read/write lock MUST be held at this point.
func (c *client) do(src Marshaler, dst Unmarshaler) error {
	c.seq++

	// get shared buffer for marshaling and unmarshaling.
	bufp, _ := bufPool.Get().(*[]byte)
	defer bufPool.Put(bufp)

	buf, err := Append((*bufp)[:0], c.seq, c.conn.Pid(), src)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = c.conn.Write(buf)
	if err != nil {
		return err
	}

	for {
		n, err := c.conn.Read(*bufp)
		if err != nil {
			return err
		}

		mr := NewMessageReader((*bufp)[:n])

		for i, msg := range mr.Each {
			if msg.Header().Type == DONE {
				// no more messages to read.
				return nil
			}

			if dst != nil {
				err := dst.UnmarshalNetlink(msg)
				if err != nil {
					return fmt.Errorf("message %d: %w", i, err)
				}
			}
		}

		if err = mr.Err(); err != nil {
			if IsACK(err) {
				// no more messages to read.
				return nil
			}

			return err
		}
	}
}
