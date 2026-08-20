// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtaddr

import (
	"errors"
	"fmt"
	"net/netip"
	"syscall"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/rt/rtroute"

	"golang.org/x/sys/unix"
)

var (
	// ErrAddrExists is returned by [AddrClient] when attempting to create an
	// [Addr] but it already exists.
	ErrAddrExists = errors.New("address exists")

	// ErrAddrNotFound is returned by [AddrClient] when attempting to remove an
	// address that does not exist.
	ErrAddrNotFound = errors.New("address not found")
)

// AddrClient implements the portion of an rtnetlink client for handling link
// address configuration.
type AddrClient struct {
	nl netlink.Client
}

// New initializes a [AddrClient] for the address subset of rtnetlink for
// exchanging messages for [Link].
//
// This client focuses only on the rtaddr subsystem, for all subsystems,
// consider using [rt.Client].
func New() (*AddrClient, error) {
	nl, err := netlink.NewClient(netlink.ROUTE, netlink.ExtendedACK(), netlink.Strict())
	if err != nil {
		return nil, err
	}

	return From(nl), nil
}

// From initializes a [AddrClient] from an existing [netlink.Client].
//
// It MUST be configured for the [netlink.ROUTE] family.
func From(nl netlink.Client) *AddrClient {
	return &AddrClient{nl: nl}
}

// Close the client and underlying socket.
func (c *AddrClient) Close() error {
	return c.nl.Close()
}

// CreateAddr configures and adds an [Addr] with flags, to the specified link
// by it's index.
//
// If it already exists, [ErrAddrExists] is returned.
func (c *AddrClient) CreateAddr(link int, flags Flags, opts ...AddrOption) error {
	if link == 0 {
		return fmt.Errorf("link index is required")
	} else if len(opts) == 0 {
		return fmt.Errorf("at least one AddrOption required")
	}

	req := &configureAddr{
		info: IfAddrMsg{
			Flags: flags,
			Link:  link,
		},
	}

	for i, opt := range opts {
		err := opt(req)
		if err != nil {
			return fmt.Errorf("option %d: %w", i, err)
		}
	}

	err := c.nl.Do(
		unix.RTM_NEWADDR,
		netlink.CREATE|netlink.EXCL,
		req,
	)
	if errors.Is(err, syscall.EEXIST) {
		return ErrAddrExists
	} else if err != nil {
		return fmt.Errorf("could not create addr: %w", err)
	}

	return nil
}

// ListAddrs retrieves a list of all the [Addr] for the specified link and
// address family.
//
// If link is zero, the list will include all links. Family may be specified
// as [rtroute.ALL] for all address families.
func (c *AddrClient) ListAddrs(link int, family rtroute.Family) (Addrs, error) {
	addrs := Addrs{}

	err := c.nl.Dump(unix.RTM_GETADDR, 0, IfAddrMsg{
		Family: family,
		Link:   link,
	}, &addrs)
	if err != nil {
		return nil, fmt.Errorf("could not list addrs: %w", err)
	}

	return addrs, nil
}

// RemoveAddr removes an [Addr] from the specified link index by it's local
// address.
//
// If it does not exist, [ErrAddrNotFound] is returned.
func (c *AddrClient) RemoveAddr(link int, local netip.Prefix) error {
	if link == 0 {
		return fmt.Errorf("link index is required")
	}

	req := &configureAddr{
		info: IfAddrMsg{
			Link: link,
		},
	}

	err := Local(local)(req)
	if err != nil {
		return err
	}

	err = c.nl.Do(unix.RTM_DELADDR, 0, req)
	if errors.Is(err, syscall.EADDRNOTAVAIL) {
		return ErrAddrNotFound
	} else if err != nil {
		return fmt.Errorf("could not remove addr: %w", err)
	}

	return nil
}
