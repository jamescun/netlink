// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package genetlink implements the Linux Kernel Generic Netlink protocol, for
// interacting with the systems network stack and related subsystems.
//
// This package contains a low-level implementation of the protocol, you should
// use a higher-level abstraction contained within a subpackage.
//
// References:
//   - linux/include/uapi/linux/netlink.h
//   - linux/include/uapi/linux/genetlink.h
//   - https://www.kernel.org/doc/html/next/userspace-api/netlink/intro.html
package genetlink

import (
	"fmt"
	"io"

	"go.jamescun.com/netlink"
)

// headerLen is the length of the Generic Netlink header in bytes.
const headerLen = 4

// GetHeader reads the Generic Netlink preamble before the Netlink message
// attributes.
func GetHeader(msg *netlink.Message) (cmd, version uint8, err error) {
	b := make([]byte, headerLen)
	_, err = io.ReadFull(msg, b)
	if err != nil {
		err = fmt.Errorf("genetlink: %w", err)
	}

	cmd = b[0]
	version = b[1]

	return
}

// SetHeader writes the Generic Netlink preamble before the Netlink message
// attributes.
func SetHeader(msg *netlink.Message, cmd, version uint8) (err error) {
	_, err = msg.Write([]byte{cmd, version, 0x00, 0x00})
	if err != nil {
		err = fmt.Errorf("genetlink: %w", err)
	}

	return
}
