// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package netlink implements the Linux Kernel Netlink protocol, for
// interacting with the systems network stack and related subsystems.
//
// This package contains a low-level implementation of the protocol, you should
// use a higher-level abstraction contained within a subpackage.
//
// References:
//   - linux/include/uapi/linux/netlink.h
//   - https://www.kernel.org/doc/html/next/userspace-api/netlink/intro.html
//   - https://www.infradead.org/~tgr/libnl/doc/core.html
package netlink

import (
	"fmt"
)

// number of bytes to align messages and attributes to.
const align = 4

// Align return length n as aligned to 4 bytes.
func Align(n int) int {
	return (n + align - 1) & ^(align - 1)
}

// Pad appends up to 4 null bytes to the end of b to align it.
func Pad(b []byte) []byte {
	const pad = "\x00\x00\x00\x00"

	n := Align(len(b)) - len(b)
	return append(b, pad[:n]...)
}

// Family configures the family of the Unix socket used for Netlink messages.
type Family int

// Constants for [Family].
const (
	ROUTE          Family = 0
	UNUSED         Family = 1
	USERSOCK       Family = 2
	FIREWALL       Family = 3
	SOCK_DIAG      Family = 4
	NFLOG          Family = 5
	XFRM           Family = 6
	SELINUX        Family = 7
	ISCSI          Family = 8
	AUDIT          Family = 9
	FIB_LOOKUP     Family = 10
	CONNECTOR      Family = 11
	NETFILTER      Family = 12
	IP6_FW         Family = 13
	DNRTMSG        Family = 14
	KOBJECT_UEVENT Family = 15
	GENERIC        Family = 16
	SCSITRANSPORT  Family = 18
	ECRYPTFS       Family = 19
	RDMA           Family = 20
	CRYPTO         Family = 21
	SMC            Family = 22
)

func (f Family) String() string {
	switch f {
	case ROUTE:
		return "ROUTE"
	case UNUSED:
		return "UNUSED"
	case USERSOCK:
		return "USERSOCK"
	case FIREWALL:
		return "FIREWALL"
	case SOCK_DIAG:
		return "SOCK_DIAG"
	case NFLOG:
		return "NFLOG"
	case XFRM:
		return "XFRM"
	case SELINUX:
		return "SELINUX"
	case ISCSI:
		return "ISCSI"
	case AUDIT:
		return "AUDIT"
	case FIB_LOOKUP:
		return "FIB_LOOKUP"
	case CONNECTOR:
		return "CONNECTOR"
	case NETFILTER:
		return "NETFILTER"
	case IP6_FW:
		return "IP6_FW"
	case DNRTMSG:
		return "DNRTMSG"
	case KOBJECT_UEVENT:
		return "KOBJECT_UEVENT"
	case GENERIC:
		return "GENERIC"
	case SCSITRANSPORT:
		return "SCSITRANSPORT"
	case ECRYPTFS:
		return "ECRYPTFS"
	case RDMA:
		return "RDMA"
	case CRYPTO:
		return "CRYPTO"
	case SMC:
		return "SMC"

	default:
		return fmt.Sprintf("Family(%d)", f)
	}
}

// Constants for builtin Netlink message types.
const (
	NOOP    uint16 = 0x01
	ERROR   uint16 = 0x02
	DONE    uint16 = 0x03
	OVERRUN uint16 = 0x04
)

// Constants for message flags.
const (
	REQUEST       uint16 = 0x01
	MULTI         uint16 = 0x02
	ACK           uint16 = 0x04
	ECHO          uint16 = 0x08
	DUMP_INTR     uint16 = 0x10
	DUMP_FILTERED uint16 = 0x20

	// GET messages.
	ROOT   uint16 = 0x100
	MATCH  uint16 = 0x200
	ATOMIC uint16 = 0x400
	DUMP   uint16 = (ROOT | MATCH)

	// NEW messages.
	REPLACE uint16 = 0x100
	EXCL    uint16 = 0x200
	CREATE  uint16 = 0x400
	APPEND  uint16 = 0x800

	// DELETE messages.
	NONREC uint16 = 0x100
	BULK   uint16 = 0x200

	// ACK messages.
	CAPPED   uint16 = 0x100
	ACK_TLVS uint16 = 0x200
)
