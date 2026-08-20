// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package rt implements rtnetlink for links, addresses, routes and network
// neighbors.
//
// Each rtnetlink subsystem is contains within a the rtlink, rtaddr, rtroute
// and rtneigh packages.
package rt

import (
	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/rt/rtaddr"
	"go.jamescun.com/netlink/rt/rtlink"
)

// Client contains embedded clients for all the rtnetlink subsystems.
//
// Note: calling [Client.Close] or Close on any subsystem client will close
// them all.
type Client struct {
	nl netlink.Client

	rtaddr.AddrClient
	rtlink.LinkClient
}

// New initializes a new [Client] containing clients for all the rtnetlink
// subsystems.
func New() (*Client, error) {
	nl, err := netlink.NewClient(netlink.ROUTE, netlink.ExtendedACK(), netlink.Strict())
	if err != nil {
		return nil, err
	}

	return &Client{
		nl: nl,

		AddrClient: *rtaddr.From(nl),
		LinkClient: *rtlink.From(nl),
	}, nil
}

// Close the rtnetlink client.
//
// Note: calling [Client.Close] or Close on any subsystem client will close
// them all.
func (c *Client) Close() error {
	return c.nl.Close()
}
