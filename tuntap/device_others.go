// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

//go:build !linux

package tuntap

import (
	"fmt"
)

type deviceOptions struct {
	flags uint16
}

func newTUN(_ string, _ ...DeviceOption) (Device, error) {
	return nil, fmt.Errorf("unsupported platform")
}

func newTAP(_ string, _ ...DeviceOption) (Device, error) {
	return nil, fmt.Errorf("unsupported platform")
}
