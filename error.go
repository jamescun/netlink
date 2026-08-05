// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Error is returned by a Netlink socket when an error occurs, containing at
// least an error code, and the original Netlink header of the request.
type Error struct {
	Code     uint32
	Original Header
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"netlink error: code: %d, type: %d, flags: %s, seq: %d, pid: %d",
		e.Code, e.Original.Type, e.Original.Flags, e.Original.Seq, e.Original.Pid,
	)
}

// UnmarshalAttributes unmarshals the body of a Netlink error.
func (e *Error) UnmarshalAttributes(attrs *AttributeReader) error {
	b := make([]byte, 4+hdrLen)
	_, err := io.ReadFull(attrs, b)
	if err != nil {
		return fmt.Errorf("netlink error: %w", err)
	}

	e.Code = binary.NativeEndian.Uint32(b)

	err = e.Original.UnmarshalBinary(b[4:])
	if err != nil {
		return fmt.Errorf("netlink error: %w", err)
	}

	return nil
}
