// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package route implements interacting with the Linux Kernel rtnetlink
// Netlink family, for managing the systems network interfaces, routing and
// networking neighbors.
//
// References:
//   - rtnetlink(7)
//   - linux/include/uapi/linux/rtnetlink.h
//   - https://www.kernel.org/doc/html/next/networking/netlink_spec/rt-link.html
package route

import (
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

func readIfInfomsg(r io.Reader) (info unix.IfInfomsg, err error) {
	b := make([]byte, 16)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return
	}

	info = *(*unix.IfInfomsg)(unsafe.Pointer(&b[0])) //nolint
	return
}

// func writeIfInfomsg(w io.Writer, info unix.IfInfomsg) error {
// 	b := *(*[16]byte)(unsafe.Pointer(&info)) //nolint
// 	_, err := w.Write(b[:])
// 	if err != nil {
// 		return fmt.Errorf("could not write IfInfomsg: %w", err)
// 	}
//
// 	return err
// }
