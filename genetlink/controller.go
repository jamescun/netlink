// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package genetlink

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

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
//   - https://www.kernel.org/doc/html/latest/netlink/specs/nlctrl.html
type Controller struct {
	client Client
}

// NewController establishes a Netlink socket connection for the
// [netlink.GENERIC] family, and returns a [Controller] client.
func NewController() (*Controller, error) {
	nl, err := NewClient("nlctrl", 2, netlink.ExtendedACK())
	if err != nil {
		return nil, err
	}

	return &Controller{
		client: nl,
	}, nil
}

// Close the Netlink socket to the [Controller] client.
func (c *Controller) Close() error {
	return c.client.Close()
}

// GetFamily is a helper to resolve the the named [Family] with an established
// [netlink.Client] without initializing a whole [Controller].
//
// The client MUST be configured for the [netlink.GENERIC] family.
//
// If the family does not exist, [ErrFamilyNotFound] is returned.
func GetFamily(nl netlink.Client, name string) (*Family, error) {
	if nl.Family() != netlink.GENERIC {
		return nil, fmt.Errorf("required GENERIC family, got %d", nl.Family())
	}

	ctrl := &Controller{client: &client{
		nl:      nl,
		family:  unix.GENL_ID_CTRL,
		version: 2,
	}}

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
	family := &Family{}

	err := c.client.Get(
		unix.CTRL_CMD_GETFAMILY,
		netlink.MarshalerFunc(func(msg netlink.MessageEncoder) error {
			err := msg.Marshal(netlink.AttributeMarshalerFunc(func(attrs *netlink.AttributeEncoder) error {
				return attrs.String(unix.CTRL_ATTR_FAMILY_NAME, name)
			}))
			if err != nil {
				return fmt.Errorf("attributes: %w", err)
			}

			return nil
		}),

		family,
	)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrFamilyNotFound
	} else if err != nil {
		return nil, err
	}

	return family, nil
}

// ListFamilies lists the available Generic Netlink families.
func (c *Controller) ListFamilies() (Families, error) {
	families := Families{}

	err := c.client.Dump(
		unix.CTRL_CMD_GETFAMILY,
		nil,
		&families,
	)
	if err != nil {
		return nil, err
	}

	return families, nil
}

// Family contains information about a registered Generic Netlink family.
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/nlctrl.html#ctrl-attrs
type Family struct {
	ID      uint16
	Name    string
	Version uint32
	HdrSize uint32
	MaxAttr uint32
	Ops     []*Op
	Groups  []*Group
}

// UnmarshalAttributes unmarshals the attributes for a Generic Netlink family.
func (f *Family) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
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

		case unix.CTRL_ATTR_OPS:
			for nested := range attr.Each {
				op := new(Op)

				err := nested.Unmarshal(op)
				if err != nil {
					return fmt.Errorf("op: %w", err)
				}

				f.Ops = append(f.Ops, op)
			}

		case unix.CTRL_ATTR_MCAST_GROUPS:
			for nested := range attr.Each {
				g := new(Group)

				err := nested.Unmarshal(g)
				if err != nil {
					return fmt.Errorf("group: %w", err)
				}

				f.Groups = append(f.Groups, g)
			}
		}
	}

	return nil
}

// UnmarshalNetlink unmarshals a Generic Netlink family from a message.
func (f *Family) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	err := msg.Unmarshal(f)
	if err != nil {
		return fmt.Errorf("attributes: %w", err)
	}

	return nil
}

// Families contains the Generic Netlink families available on the system.
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/nlctrl.html#ctrl-attrs
type Families []*Family

// UnmarshalNetlink unmarshals a Generic Netlink family dump from a message.
func (fs *Families) UnmarshalNetlink(msg netlink.MessageDecoder) error {
	f := &Family{}

	err := f.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	*fs = append(*fs, f)

	return nil
}

// Group is one of the multicast groups a [Family] has defined.
//
// References:
//   - https://www.kernel.org/doc/html/latest/netlink/specs/nlctrl.html#mcast-group-attrs
type Group struct {
	ID   uint32
	Name string
}

// UnmarshalAttributes unmarshals a multicast group contained within [Family].
func (g *Group) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.CTRL_ATTR_MCAST_GRP_ID:
			g.ID = attr.Uint32()
		case unix.CTRL_ATTR_MCAST_GRP_NAME:
			g.Name = attr.String()
		}
	}

	return nil
}

// Op is one of the Generic Netlink operations a [Family] has defined.
//
// References:
//   - linux/include/uapi/linux/genetlink.h
type Op struct {
	ID    uint32
	Flags OpFlags
}

// UnmarshalAttributes unmarshals a single [Op] for a [Family].
func (o *Op) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.CTRL_ATTR_OP_ID:
			o.ID = attr.Uint32()
		case unix.CTRL_ATTR_OP_FLAGS:
			o.Flags = OpFlags(attr.Uint32())
		}
	}

	return nil
}

// OpFlags defines the permissions required for an [Op].
//
// References:
//   - linux/include/uapi/linux/genetlink.h
type OpFlags uint32

// Constants for [Op].
const (
	ADMIN_PERM OpFlags = 1 << iota
	CMD_CAP_DO
	CMD_CAP_DUMP
	CMD_CAP_HASPOL
	UNS_ADMIN_PERM
)

// opFlagNames is bit shifted through by [OpFlags.String] to build a
// stringified representation.
var opFlagNames = []string{
	"ADMIN_PERM",
	"CMD_CAP_DO",
	"CMD_CAP_DUMP",
	"CMD_CAP_HASPOL",
	"UNS_ADMIN_PERM",
}

func (o OpFlags) String() string {
	if o == 0 {
		return "NONE"
	}

	var s strings.Builder

	for i, name := range opFlagNames {
		if o&(1<<i) != 0 {
			s.WriteByte(' ')
			s.WriteString(name)
		}
	}

	if s.Len() != 0 {
		return s.String()[1:]
	}

	return ""
}
