// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

//go:build !linux

package netlink

import (
	"fmt"
)

func dial(_ Family, _ ...ConnOption) (Conn, error) {
	return nil, fmt.Errorf("unsupported platform")
}
