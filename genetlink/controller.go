// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package genetlink

import (
	"errors"
	"fmt"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// ErrFamilyNotFound is returned by [Controller.GetFamily] when the requested
// Generic Netlink family does not exist.
var ErrFamilyNotFound = errors.New("family not found")

// Controller implements the API for the Generic Netlink Controller, used to
// register and discover Generic Netlink families available on a system.
//
// References:
//   - https://www.kernel.org/doc/html/v6.12/networking/netlink_spec/nlctrl.html
type Controller struct {
	client netlink.Client
}

// NewController initializes a new [Controller] implementation for the given
// [netlink.Client].
//
// It MUST be configured for the [netlink.GENERIC] family.
func NewController(client netlink.Client) (*Controller, error) {
	if client.Family() != netlink.GENERIC {
		return nil, fmt.Errorf("controller: required GENERIC netlink family, got %s", client.Family())
	}

	return &Controller{
		client: client,
	}, nil
}

// GetFamily is a helper to resolve the the named [Family] with an established
// [netlink.Client] without initializing a whole [Controller].
//
// The client MUST be configured for the [netlink.GENERIC] family.
//
// If the family does not exist, [ErrFamilyNotFound] is returned.
func GetFamily(client netlink.Client, name string) (*Family, error) {
	ctrl, err := NewController(client)
	if err != nil {
		return nil, err
	}

	family, err := ctrl.GetFamily(name)
	if err != nil {
		return nil, fmt.Errorf("genetlink: %w", err)
	}

	return family, nil
}

// GetFamily returns details of the named Generic Netlink family.
//
// If the family does not exist, [ErrFamilyNotFound] is returned.
func (c *Controller) GetFamily(name string) (*Family, error) {
	f := &Family{}

	err := c.client.Do(getFamily{name: name}, f)
	if _, ok := err.(*netlink.Error); ok {
		return nil, ErrFamilyNotFound
	} else if err != nil {
		return nil, err
	}

	return f, nil
}

// ListFamilies lists the available Generic Netlink families.
func (c *Controller) ListFamilies() (Families, error) {
	fs := Families{}

	err := c.client.Do(getFamily{}, &fs)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// Family contains information about a registered Generic Netlink family.
//
// References:
//   - https://www.kernel.org/doc/html/v6.12/networking/netlink_spec/nlctrl.html#id5
type Family struct {
	ID          uint16
	Name        string
	Version     uint32
	HdrSize     uint32
	MaxAttr     uint32
	McastGroups []McastGroup
}

// UnmarshalAttributes unmarshals a Generic Netlink family from attributes.
func (f *Family) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.CTRL_ATTR_FAMILY_ID:
			f.ID = attr.Uint16()
		case unix.CTRL_ATTR_FAMILY_NAME:
			f.Name = attr.String()
		case unix.CTRL_ATTR_VERSION:
			f.Version = attr.Uint32()
		case unix.CTRL_ATTR_HDRSIZE:
			f.HdrSize = attr.Uint32()
		case unix.CTRL_ATTR_MAXATTR:
			f.MaxAttr = attr.Uint32()

		case unix.CTRL_ATTR_MCAST_GROUPS:
			for nested := range attr.Array {
				mg := McastGroup{}

				err := mg.UnmarshalAttributes(nested)
				if err != nil {
					return err
				}

				f.McastGroups = append(f.McastGroups, mg)
			}
		}
	}

	return nil
}

// UnmarshalNetlink unmarshals a single Generic Netlink family.
func (f *Family) UnmarshalNetlink(msg *netlink.Message) error {
	_, _, err := GetHeader(msg)
	if err != nil {
		return err
	}

	err = msg.Unmarshal(f)
	if err != nil {
		return err
	}

	return nil
}

// Families contains the Generic Netlink families available on the system.
//
// References:
//   - https://www.kernel.org/doc/html/v6.12/networking/netlink_spec/nlctrl.html#id5
type Families []Family

// UnmarshalNetlink unmarshals Generic Netlink families.
func (fs *Families) UnmarshalNetlink(msg *netlink.Message) error {
	if msg.Type == netlink.DONE {
		// received the last family.
		return nil
	}

	f := Family{}

	err := f.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	*fs = append(*fs, f)

	return nil
}

// McastGroup is one of the multicast groups a [Family] has defined.
//
// References:
//   - https://docs.kernel.org/netlink/specs/nlctrl.html#nlctrl-attribute-set-mcast-group-attrs
//   - https://www.kernel.org/doc/html/v6.7/userspace-api/netlink/intro.html#multicast-notifications
type McastGroup struct {
	ID   uint32
	Name string
}

// UnmarshalAttributes unmarshals a multicast group contained within [Family].
func (mg *McastGroup) UnmarshalAttributes(attrs *netlink.AttributeReader) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.CTRL_ATTR_MCAST_GRP_ID:
			mg.ID = attr.Uint32()
		case unix.CTRL_ATTR_MCAST_GRP_NAME:
			mg.Name = attr.String()
		}
	}

	return nil
}

type getFamily struct {
	name string
}

func (gf getFamily) MarshalAttributes(attrs *netlink.AttributeWriter) error {
	if gf.name != "" {
		err := attrs.AddString(unix.CTRL_ATTR_FAMILY_NAME, gf.name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gf getFamily) MarshalNetlink(msg *netlink.Message) error {
	msg.Type = 0x10
	msg.Flags = netlink.REQUEST | netlink.ACK

	if gf.name == "" {
		msg.Flags |= netlink.DUMP
	}

	err := SetHeader(msg, 0x03, 0x01)
	if err != nil {
		return err
	}

	err = msg.Marshal(gf)
	if err != nil {
		return err
	}

	return nil
}
