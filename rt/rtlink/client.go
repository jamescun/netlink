// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"errors"
	"fmt"
	"syscall"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

var (
	// ErrLinkNotFound is returned by [Client] when attempting to get or
	// configure a [Link] by index or name but it does not exist.
	ErrLinkNotFound = errors.New("link not found")

	// ErrLinkExists is returned by [Client] when attempting to create a [Link]
	// but it already exists.
	ErrLinkExists = errors.New("link exists")
)

// Client implements the portion of an rtnetlink client for handling link
// configuration.
type Client struct {
	nl netlink.Client
}

// New initializes a [Client] for the subset of rtnetlink for exchanging
// messages for [Link].
func New() (*Client, error) {
	nl, err := netlink.NewClient(netlink.ROUTE, netlink.ExtendedACK(), netlink.Strict())
	if err != nil {
		return nil, err
	}

	return From(nl), nil
}

// From initializes a [Client] from an existing [netlink.Client].
//
// It MUST be configured for the [netlink.ROUTE] family.
func From(nl netlink.Client) *Client {
	return &Client{nl: nl}
}

// Close the client and underlying socket.
func (c *Client) Close() error {
	return c.nl.Close()
}

// ConfigureLink is used to configure an existing [Link].
//
// It is given the index for the link, and one-or-more [LinkOption] to apply
// those configuration options.
//
// If the link does not exist, [ErrLinkNotFound] is returned.
func (c *Client) ConfigureLink(index int, opts ...LinkOption) error {
	req := &configureLink{
		info: IfInfoMsg{
			Index: index,
		},
	}

	for i, opt := range opts {
		err := opt(req)
		if err != nil {
			return fmt.Errorf("option %d: %w", i, err)
		}
	}

	err := c.nl.Do(unix.RTM_SETLINK, 0, req)
	if errors.Is(err, syscall.ENODEV) {
		return ErrLinkNotFound
	} else if err != nil {
		return fmt.Errorf("could not configure link: %w", err)
	}

	return nil
}

// CreateLink is used to create and configure a new [Link].
//
// It is given the name for the new link, the driver-specific device
// configuration and optional [LinkOption] to set during creation.
//
// If the link does not require additional configuration, or is configured
// through another interface, [Generic] can be used to specify the device
// driver name.
//
// If the link already exists, [ErrLinkExists] is returned.
func (c *Client) CreateLink(name string, device Device, opts ...LinkOption) error {
	if name == "" {
		return fmt.Errorf("name is required")
	} else if len(name) > 15 {
		return fmt.Errorf("name exceeds 15 bytes, got %d", len(name))
	} else if device == nil {
		return fmt.Errorf("device to create is required")
	}

	req := &configureLink{
		linkInfo: &Info{
			Kind: device.DeviceKind(),
			Data: device,
		},
	}

	for i, opt := range opts {
		err := opt(req)
		if err != nil {
			return fmt.Errorf("option %d: %w", i, err)
		}
	}

	// don't allow [Name] to override the one given to this method.
	req.name = new(name)

	err := c.nl.Do(unix.RTM_NEWLINK, netlink.CREATE|netlink.EXCL, req)
	if errors.Is(err, syscall.EEXIST) {
		return ErrLinkExists
	} else if err != nil {
		return fmt.Errorf("could not create link: %w", err)
	}

	return nil
}

// GetLinkByIndex retrieves a [Link] by it's index.
//
// If the link does not exist, [ErrLinkNotFound] is returned.
func (c *Client) GetLinkByIndex(index int) (*Link, error) {
	link := &Link{}

	err := c.nl.Get(unix.RTM_GETLINK, 0, &IfInfoMsg{Index: index}, link)
	if errors.Is(err, syscall.ENODEV) {
		return nil, ErrLinkNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get link: %w", err)
	}

	return link, nil
}

// GetLinkByName retrieves a [Link] by it's name.
//
// If the link does not exist, [ErrLinkNotFound] is returned.
func (c *Client) GetLinkByName(name string) (*Link, error) {
	link := &Link{}

	err := c.nl.Get(unix.RTM_GETLINK, 0, netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
		err := msg.MarshalBytes(&IfInfoMsg{})
		if err != nil {
			return fmt.Errorf("ifinfomsg: %w", err)
		}

		err = msg.Marshal(netlink.AttributeMarshalerFunc(func(attrs *netlink.AttributeEncoder) error {
			return attrs.String(unix.IFLA_IFNAME, name)
		}))
		if err != nil {
			return fmt.Errorf("attributes: %w", err)
		}

		return nil
	}), link)
	if errors.Is(err, syscall.ENODEV) {
		return nil, ErrLinkNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get link: %w", err)
	}

	return link, nil
}

// ListLinks retrieves a list of [Link], optionally filtered by family and
// some attribute types (that may be expensive to retrieve)
func (c *Client) ListLinks(family Family, filter Filter) (Links, error) {
	links := Links{}

	err := c.nl.Dump(unix.RTM_GETLINK, 0, netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
		err := msg.MarshalBytes(IfInfoMsg{Family: family})
		if err != nil {
			return fmt.Errorf("ifinfomsg: %w", err)
		}

		err = msg.Marshal(netlink.AttributeMarshalerFunc(func(attrs *netlink.AttributeEncoder) error {
			attrs.Uint32(unix.IFLA_EXT_MASK, uint32(filter))
			return nil
		}))
		if err != nil {
			return fmt.Errorf("attributes: %w", err)
		}

		return nil
	}), &links)
	if errors.Is(err, syscall.ENODEV) {
		return nil, ErrLinkNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get link: %w", err)
	}

	return links, nil
}

// RemoveLink removes a [Link] by it's index.
//
// If the link does not exist, [ErrLinkNotFound] is returned.
func (c *Client) RemoveLink(index int) error {
	err := c.nl.Do(unix.RTM_DELLINK, 0, &IfInfoMsg{Index: index})
	if errors.Is(err, syscall.ENODEV) {
		return ErrLinkNotFound
	} else if err != nil {
		return fmt.Errorf("could not remove link: %w", err)
	}

	return nil
}
