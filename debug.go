// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"io"
)

type dumper struct {
	Conn

	r io.Writer
	w io.Writer
}

// Dumper initializes a [Conn], wrapping another [Conn], that writes the
// contents of each read or write to an [io.Writer], for later testing or
// development.
//
// Either writer may be nil, to only capture one half of the exchange.
func Dumper(r, w io.Writer, conn Conn) Conn {
	return &dumper{
		Conn: conn,
		r:    r,
		w:    w,
	}
}

func (d *dumper) Read(b []byte) (int, error) {
	n, err := d.Conn.Read(b)
	if err != nil {
		return n, err
	}

	if d.r != nil {
		d.r.Write(b[:n]) //nolint
	}

	return n, nil
}

func (d *dumper) Write(b []byte) (int, error) {
	n, err := d.Conn.Write(b)
	if err != nil {
		return n, err
	}

	if d.w != nil {
		d.w.Write(b) //nolint
	}

	return n, nil
}
