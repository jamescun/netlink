// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtaddr

import (
	"fmt"
	"net/netip"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/rt/rtroute"

	"golang.org/x/sys/unix"
)

// configureAddr is built through user-specified [AddrOption] configure an
// address for a link.
type configureAddr struct {
	info      IfAddrMsg
	address   *netip.Addr
	local     *netip.Addr
	label     *string
	cacheInfo *CacheInfo
	delete    bool
}

// AddrOption is a function used to configure an [Addr] being created or
// removed.
type AddrOption func(*configureAddr) error

// Address is an [AddrOption] to configure the Network Address and Prefix of an
// [Addr] being created or removed.
//
// Note: the address specified is the prefix address, it makes no difference
// for normally configured broadcast-based interfaces; however for
// POINT_TO_POINT interfaces it is the destination address, and the local
// address is specified with [Local].
func Address(address netip.Prefix) AddrOption {
	return func(addr *configureAddr) error {
		addr.info.PrefixLen = uint8(address.Bits()) //nolint
		addr.address = new(address.Addr())

		if address.Addr().Is6() {
			addr.info.Family = rtroute.INET6
		} else {
			addr.info.Family = rtroute.INET
		}

		return nil
	}
}

// Label is an [AddrOption] to configure the administrative label of an [Addr]
// being created.
func Label(label string) AddrOption {
	return func(addr *configureAddr) error {
		addr.label = new(label)
		return nil
	}
}

// Local is an [AddrOption] to configure the local Network Address and Prefix
// of an [Addr] being created or removed.
func Local(local netip.Prefix) AddrOption {
	return func(addr *configureAddr) error {
		addr.info.PrefixLen = uint8(local.Bits()) //nolint
		addr.local = new(local.Addr())

		if local.Addr().Is6() {
			addr.info.Family = rtroute.INET6
		} else {
			addr.info.Family = rtroute.INET
		}

		return nil
	}
}

// Validity is an [AddrOption] to configure the cache information for an [Addr]
// being created. If not specified, [Forever] will be the default.
func Validity(info *CacheInfo) AddrOption {
	return func(addr *configureAddr) error {
		if info == nil {
			return fmt.Errorf("CacheInfo is required")
		}

		addr.cacheInfo = info
		return nil
	}
}

func (c *configureAddr) MarshalAttributes(attrs *netlink.AttributeEncoder) error {
	if c.address != nil {
		attrs.Addr(unix.IFA_ADDRESS, *c.address)
	}

	if c.local != nil {
		attrs.Addr(unix.IFA_LOCAL, *c.local)
	}

	if c.delete {
		// no more options can be applied to delete messages.
		return nil
	}

	if c.label != nil {
		err := attrs.String(unix.IFA_LABEL, *c.label)
		if err != nil {
			return fmt.Errorf("label: %w", err)
		}
	}

	if c.cacheInfo != nil {
		err := attrs.MarshalBytes(unix.IFA_CACHEINFO, c.cacheInfo)
		if err != nil {
			return fmt.Errorf("cache info: %w", err)
		}
	}

	return nil
}

func (c *configureAddr) MarshalNetlink(msg netlink.MessageEncoder) error {
	err := msg.MarshalBytes(c.info)
	if err != nil {
		return fmt.Errorf("ifaddrmsg: %w", err)
	}

	err = msg.Marshal(c)
	if err != nil {
		return err
	}

	return nil
}
