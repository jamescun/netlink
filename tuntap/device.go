// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package tuntap implements creating TUN and TAP device using the TUNTAP
// device subsystem.
//
// While TUNTAP devices do not directly make use of netlink, it is the only
// device type that isn't created and configured through rtnetlink. This
// package bridges that gap, and is built in such a way as to be fully
// compatible with the rtnetlink package.
package tuntap

import (
	"time"
)

// Device is either a TUN device for receiving raw layer-3 packets, or a TAP
// device for receiving raw layer-2 packets, integrated with the Go networking
// polling infrastructure.
type Device interface {
	// Close the device, and unblock any in-flight read or write operation.
	Close() error

	// Name returns the name assigned to this link.
	//
	// This is safe to call even after the device has been closed, however if
	// the device has not been persisted, it may be reused.
	Name() string

	// Index returns the unique index of the link associated with this device,
	// which can be used with the rtnetlink package to configure it.
	//
	// This is safe to call even after the device has been closed, however if
	// the device has not been persisted, it may be reused.
	Index() int

	// Read data from the TUNTAP device, either a layer-3 packet for a TUN
	// device, or a layer-2 frame for a TAP device.
	//
	// Unless the [NoPacketInfo] device option is set, it will include the
	// TUNTAP frame header at the beginning of the bytes.
	Read([]byte) (int, error)

	// Write data to the TUNTAP device, either a layer-3 packet for a TUN
	// device, or a layer-2 frame for a TAP device.
	Write([]byte) (int, error)

	// SetDeadline sets the read and write deadlines associated with the
	// connection. It is equivalent to calling both SetReadDeadline and
	// SetWriteDeadline.
	SetDeadline(time.Time) error

	// SetReadDeadline sets the deadline for future Read calls and any
	// currently-blocked Read call.
	//
	// A zero value for t means Read will not time out.
	SetReadDeadline(time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls and any
	// currently-blocked Write call.
	//
	// Even if write times out, it may return n > 0, indicating that some of
	// the data was successfully written.
	//
	// A zero value for t means Write will not time out.
	SetWriteDeadline(time.Time) error
}

// NewTUN initializes a new layer-3 TUN device of the given name, optionally
// configured using the given [DeviceOption].
//
// The name must not be in-use already, unless the [MultiQueue] device option
// is specified.
func NewTUN(name string, opts ...DeviceOption) (Device, error) {
	return newTUN(name, opts...)
}

// NewTAP initializes a new layer-2 TAP device of the given name, optionally
// configured using the given [DeviceOption].
//
// The name must not be in-use already, unless the [MultiQueue] device option
// is specified.
func NewTAP(name string, opts ...DeviceOption) (Device, error) {
	return newTAP(name, opts...)
}

// DeviceOption is a function that allows for additional configuration of a
// TUNTAP device when being created.
type DeviceOption func(*deviceOptions)
