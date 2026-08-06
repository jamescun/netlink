// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"fmt"
)

// Error is a Netlink error message received when a fault occurs, containing
// the error code, header of original message, and optionally extended
// attributes.
type Error struct {
	// Code is the error code received.
	Code int

	// Request contains the header of the original request message.
	Request Header

	// Inner contains the actual [syscall.Errno].
	Inner error
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"netlink: code=%d seq=%d pid=%d: %s",
		e.Code, e.Request.Seq, e.Request.Pid, e.Inner,
	)
}

func (e *Error) Unwrap() error {
	return e.Inner
}
