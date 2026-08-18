// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"fmt"
	"math"
	"net"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// configureLink is built through user-specified [LinkOption] to create,
// configure and delete a [Link].
//
// References:
//   - linux/net/core/rtnetlink.c nla_policy
type configureLink struct {
	info      IfInfoMsg
	address   net.HardwareAddr
	broadcast net.HardwareAddr
	name      *string
	mtu       *uint32
	master    *uint32
	txQlen    *uint32
	weight    *uint32
	linkMode  *uint8
	linkInfo  *Info
	netNsPid  *uint32
	alias     *string
	netNsFd   *uint32
	carrier   *uint8
}

// LinkOption is a function used to configure all [Link] types.
type LinkOption func(*configureLink) error

// Address is a [LinkOption] to set the physical address of a [Link].
func Address(address net.HardwareAddr) LinkOption {
	return func(link *configureLink) error {
		link.address = address
		return nil
	}
}

// Alias is a [LinkOption] to set the name alias of a [Link].
//
// If the alias is empty, the existing alias will be removed.
func Alias(alias string) LinkOption {
	return func(link *configureLink) error {
		if len(alias) > 15 {
			return fmt.Errorf("alias exceeds 15 bytes, got %d", len(alias))
		}

		link.alias = new(alias)
		return nil
	}
}

// Broadcast is a [LinkOption] to set the broadcast address of a [Link].
func Broadcast(broadcast net.HardwareAddr) LinkOption {
	return func(link *configureLink) error {
		link.broadcast = broadcast
		return nil
	}
}

// Carrier is a [LinkOption] to configure the carrier of a [Link].
func Carrier(carrier uint8) LinkOption {
	return func(link *configureLink) error {
		link.carrier = new(carrier)
		return nil
	}
}

// Down is a [LinkOption] to unset the [UP] flag on a [Link].
func Down(link *configureLink) error {
	link.info.Flags ^= UP
	link.info.Changed |= UP
	return nil
}

// LinkMode is a [LinkOption] to configure the [LinkMode] of a [Link].
func LinkMode(linkMode Mode) LinkOption {
	return func(link *configureLink) error {
		link.linkMode = new(uint8(linkMode))
		return nil
	}
}

// Master is a [LinkOption] to configure the master link of a [Link], such as
// adding a link to a [Bridge] device.
func Master(index int) LinkOption {
	return func(link *configureLink) error {
		if index < 0 || index > math.MaxUint32 {
			return fmt.Errorf("master exceeds uint32, got %d", index)
		}

		link.master = new(uint32(index))
		return nil
	}
}

// MTU is a [LinkOption] to configure the Maximum Transmission Unit of a
// [Link].
func MTU(mtu uint32) LinkOption {
	return func(link *configureLink) error {
		link.mtu = new(mtu)
		return nil
	}
}

// NetNsFd is a [LinkOption] to configure the network namespace of a [Link]
// based on an open file descriptor from that namespace.
//
// This option is mutually exclusive with [NetNsPid].
func NetNsFd(fd int) LinkOption {
	return func(link *configureLink) error {
		if fd < 0 || fd > math.MaxUint32 {
			return fmt.Errorf("fd exceeds uint32, got %d", fd)
		}

		if link.netNsPid != nil {
			return fmt.Errorf("net ns pid is already set")
		}

		link.netNsFd = new(uint32(fd))
		return nil
	}
}

// NetNsPid is a [LinkOption] to configure the network namespace of a [Link]
// based on the process id of another running process.
//
// This option is mutually exclusive with [NetNsFd].
func NetNsPid(pid int) LinkOption {
	return func(link *configureLink) error {
		if pid < 0 || pid > math.MaxUint32 {
			return fmt.Errorf("pid exceeds uint32, got %d", pid)
		}

		if link.netNsFd != nil {
			return fmt.Errorf("net ns fd is already set")
		}

		link.netNsPid = new(uint32(pid))
		return nil
	}
}

// TxQlen is a [LinkOption] to configure the size of the transmission queue of
// a [Link].
func TxQlen(txQlen uint32) LinkOption {
	return func(link *configureLink) error {
		link.txQlen = new(txQlen)
		return nil
	}
}

// Up is a [LinkOption] to set the [UP] flag on a [Link].
func Up(link *configureLink) error {
	link.info.Flags |= UP
	link.info.Changed |= UP
	return nil
}

// Weight is a [LinkOption] to configure the balancing weight of a [Link].
func Weight(weight uint32) LinkOption {
	return func(link *configureLink) error {
		link.weight = new(weight)
		return nil
	}
}

func (l *configureLink) MarshalAttributes(attrs *netlink.AttributeEncoder) error {
	if l.name != nil {
		err := attrs.String(unix.IFLA_IFNAME, *l.name)
		if err != nil {
			return fmt.Errorf("name: %w", err)
		}
	}

	if l.address != nil {
		attrs.HardwareAddr(unix.IFLA_ADDRESS, l.address)
	}

	if l.broadcast != nil {
		attrs.HardwareAddr(unix.IFLA_BROADCAST, l.broadcast)
	}

	if l.mtu != nil {
		attrs.Uint32(unix.IFLA_MTU, *l.mtu)
	}

	if l.master != nil {
		attrs.Uint32(unix.IFLA_MASTER, *l.master)
	}

	if l.txQlen != nil {
		attrs.Uint32(unix.IFLA_TXQLEN, *l.txQlen)
	}

	if l.weight != nil {
		attrs.Uint32(unix.IFLA_WEIGHT, *l.weight)
	}

	if l.linkMode != nil {
		attrs.Uint8(unix.IFLA_LINKMODE, *l.linkMode)
	}

	if l.linkInfo != nil {
		err := attrs.Marshal(unix.IFLA_LINKINFO, l.linkInfo)
		if err != nil {
			return fmt.Errorf("info: %w", err)
		}
	}

	if l.netNsPid != nil {
		attrs.Uint32(unix.IFLA_NET_NS_PID, *l.netNsPid)
	}

	if l.alias != nil {
		err := attrs.String(unix.IFLA_IFALIAS, *l.alias)
		if err != nil {
			return fmt.Errorf("alias: %w", err)
		}
	}

	if l.netNsFd != nil {
		attrs.Uint32(unix.IFLA_NET_NS_FD, *l.netNsFd)
	}

	if l.carrier != nil {
		attrs.Uint8(unix.IFLA_CARRIER, *l.carrier)
	}

	return nil
}

func (l *configureLink) MarshalNetlink(msg netlink.MessageEncoder) error {
	err := msg.MarshalBytes(l.info)
	if err != nil {
		return fmt.Errorf("ifinfomsg: %w", err)
	}

	err = msg.Marshal(l)
	if err != nil {
		return err
	}

	return nil
}

// genericDevice is a [Device] that only contains it's kind and takes no
// further attributes, for use with [Client.CreateLink].
type genericDevice struct {
	kind string
}

// Generic initializes a generic [Device] for device types that take no further
// configuration outside the general-purpose [LinkOption].
func Generic(kind string) Device {
	return &genericDevice{kind: kind}
}

func (g genericDevice) DeviceKind() string { return g.kind }

func (g genericDevice) MarshalAttributes(attrs *netlink.AttributeEncoder) error {
	err := attrs.String(unix.IFLA_INFO_KIND, g.kind)
	if err != nil {
		return fmt.Errorf("kind: %w", err)
	}

	return nil
}
