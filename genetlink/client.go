// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package genetlink

import (
	"fmt"

	"go.jamescun.com/netlink"
)

// Client is a wrapper around a [netlink.Client] for exchanging client requests
// and responses using Generic Netlink, where all messages will automatically
// be configured for the Generic Netlink header.
//
// It is safe for concurrent use, the client will only ever have one request
// and response exchange in-flight at once.
type Client interface {
	// Close the client, underlying socket, and prevent further client
	// exchanges.
	Close() error

	// Seq returns the last sequence number used by the client.
	Seq() uint32

	// Family returns the id of the Generic Netlink family resolved for the
	// family name this client was opened with.
	Family() uint16

	// FamilyName returns the name of the Generic Netlink family this client
	// was opened with.
	FamilyName() string

	// Version returns the Generic Netlink version this client was opened with.
	Version() uint8

	// Do executes a Generic Netlink request, marshaling the request from an
	// [Marshaler], and does not expect a response other than an ERROR
	// message type or an acknowledgement.
	//
	// The message will automatically have the Generic Netlink header added
	// with the given command, and have the REQUEST, ACK and any additional
	// flags set.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Do(cmd uint8, flags uint16, src netlink.Marshaler) error

	// Dump executes a Generic Netlink request, marshaling the request from a
	// [Marshaler], which may be nil, and unmarshaling the response(s) to an
	// [Unmarshaler].
	//
	// The message will automatically have the Generic Netlink header added
	// with the given command, and have the REQUEST, DUMP and any additional
	// flags set.
	//
	// If the response contains multiple messages, dst will be unmarshaled to
	// repeatedly until all messages have been consumed, not including the
	// DONE message type.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Dump(cmd uint8, flags uint16, src netlink.Marshaler, dst netlink.Unmarshaler) error

	// Get executes a Generic Netlink request, marshaling the request from a
	// [Marshaler], which may be nil, and unmarshaling the response to an
	// [Unmarshaler].
	//
	// The message will automatically have the Generic Netlink header added
	// with the given command, and have the REQUEST, ACK and any additional
	// flags set.
	//
	// On non-zero ERROR message types, the error will be unmarshaled
	// automatically and returned as an [Error] type.
	Get(cmd uint8, flags uint16, src netlink.Marshaler, dst netlink.Unmarshaler) error
}

type client struct {
	nl      netlink.Client
	family  uint16
	name    string
	version uint8
}

// NewClient establishes a Netlink socket connection, configured for the given
// Generic Netlink family and version, and creates a [Client] for exchanging
// messages with request-response semantics.
//
// It may optionally be given [ConnOption] to configure the underlying socket.
func NewClient(family string, version uint8, opts ...netlink.ConnOption) (Client, error) {
	nl, err := netlink.NewClient(netlink.GENERIC, opts...)
	if err != nil {
		return nil, err
	}

	f, err := GetFamily(nl, family)
	if err != nil {
		return nil, err
	}

	return &client{
		nl:      nl,
		family:  f.ID,
		name:    family,
		version: version,
	}, nil
}

func (c *client) Close() error {
	return c.nl.Close()
}

func (c *client) Family() uint16 {
	return c.family
}

func (c *client) FamilyName() string {
	return c.name
}

func (c *client) Version() uint8 {
	return c.version
}

func (c *client) Seq() uint32 {
	return c.nl.Seq()
}

func (c *client) Do(cmd uint8, flags uint16, src netlink.Marshaler) error {
	if src == nil {
		return fmt.Errorf("Marshaler is nil")
	}

	return c.nl.Do(c.family, flags, netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
		err := msg.MarshalBytes(&Header{
			Cmd:     cmd,
			Version: c.version,
		})
		if err != nil {
			return fmt.Errorf("genetlink: %w", err)
		}

		return src.MarshalNetlink(msg)
	}))
}

func (c *client) Dump(cmd uint8, flags uint16, src netlink.Marshaler, dst netlink.Unmarshaler) error {
	return c.nl.Dump(c.family, flags, netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
		err := msg.MarshalBytes(&Header{
			Cmd:     cmd,
			Version: c.version,
		})
		if err != nil {
			return fmt.Errorf("genetlink: %w", err)
		}

		if src != nil {
			return src.MarshalNetlink(msg)
		}

		return nil
	}), netlink.UnmarshalerFunc(func(msg netlink.MessageDecoder) error {
		err := msg.UnmarshalBytes(netlink.Discard(4))
		if err != nil {
			return fmt.Errorf("genetlink: %w", err)
		}

		if dst != nil {
			return dst.UnmarshalNetlink(msg)
		}

		return nil
	}))
}

func (c *client) Get(cmd uint8, flags uint16, src netlink.Marshaler, dst netlink.Unmarshaler) error {
	return c.nl.Get(c.family, flags, netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
		err := msg.MarshalBytes(&Header{
			Cmd:     cmd,
			Version: c.version,
		})
		if err != nil {
			return fmt.Errorf("genetlink: %w", err)
		}

		if src != nil {
			return src.MarshalNetlink(msg)
		}

		return nil
	}), netlink.UnmarshalerFunc(func(msg netlink.MessageDecoder) error {
		err := msg.UnmarshalBytes(netlink.Discard(4))
		if err != nil {
			return fmt.Errorf("genetlink: %w", err)
		}

		if dst != nil {
			return dst.UnmarshalNetlink(msg)
		}

		return nil
	}))
}
