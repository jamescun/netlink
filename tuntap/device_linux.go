// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package tuntap

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type deviceOptions struct {
	mtu   uint32
	flags uint16
}

// MTU is a [DeviceOption] that configures the Maximum Transmission Unit (MTU)
// for a TUNTAP device.
//
// If not set, it will default to 1500.
func MTU(mtu uint32) DeviceOption {
	return func(options *deviceOptions) {
		options.mtu = mtu
	}
}

// MultiQueue is a [DeviceOption] that configures a TUNTAP device to support
// multiple queues, in practice this means a device may be initialized multiple
// times for concurrency and performance.
func MultiQueue() DeviceOption {
	return func(options *deviceOptions) {
		options.flags |= unix.IFF_MULTI_QUEUE
	}
}

// NoPacketInfo is a [DeviceOption] that configures a TUNTAP device to not
// include the TUNTAP frame header in each read.
func NoPacketInfo() DeviceOption {
	return func(options *deviceOptions) {
		options.flags |= unix.IFF_NO_PI
	}
}

// Persist is a [DeviceOption] that configures a TUNTAP device to persist
// beyond the lifetime of the process(es) that are using it.
func Persist() DeviceOption {
	return func(options *deviceOptions) {
		options.flags |= unix.IFF_PERSIST
	}
}

type device struct {
	*os.File

	name  string
	index int
}

func newTUN(name string, opts ...DeviceOption) (Device, error) {
	options := &deviceOptions{
		flags: unix.IFF_TUN,
	}

	for _, opt := range opts {
		opt(options)
	}

	return newDevice(name, options)
}

func newTAP(name string, opts ...DeviceOption) (Device, error) {
	options := &deviceOptions{
		flags: unix.IFF_TAP,
	}

	for _, opt := range opts {
		opt(options)
	}

	return newDevice(name, options)
}

func newDevice(name string, options *deviceOptions) (Device, error) {
	if len(name) >= unix.IFNAMSIZ {
		return nil, fmt.Errorf("name cannot exceed 15 bytes, got %d", len(name))
	}

	req, err := unix.NewIfreq(name)
	if err != nil {
		return nil, fmt.Errorf("ifreq: %w", err)
	}

	req.SetUint16(options.flags)

	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	err = unix.IoctlIfreq(fd, unix.TUNSETIFF, req)
	if err != nil {
		unix.Close(fd) //nolint
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		unix.Close(fd) //nolint
		return nil, fmt.Errorf("setnonblock: %w", err)
	}

	index, err := getLinkIndex(name)
	if err != nil {
		unix.Close(fd) //nolint
		return nil, fmt.Errorf("index: %w", err)
	}

	if options.mtu != 0 {
		err = setLinkMTU(name, options.mtu)
		if err != nil {
			unix.Close(fd) //nolint
			return nil, fmt.Errorf("mtu: %w", err)
		}
	}

	return &device{
		File: os.NewFile(uintptr(fd), name),

		name:  name,
		index: index,
	}, nil
}

func (d *device) Name() string { return d.name }
func (d *device) Index() int   { return d.index }

func (d *device) String() string {
	return "TUNTAP(" + d.name + ")"
}

// getLinkIndex uses the ioctl SIOCGIFINDEX call to get the link index of the
// TUNTAP device so it can be used with the rtnetlink package.
func getLinkIndex(name string) (int, error) {
	req, err := unix.NewIfreq(name)
	if err != nil {
		return 0, fmt.Errorf("ifreq: %w", err)
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return 0, fmt.Errorf("socket: %w", err)
	}
	defer unix.Close(fd)

	err = unix.IoctlIfreq(fd, unix.SIOCGIFINDEX, req)
	if err != nil {
		return 0, fmt.Errorf("ioctl: %w", err)
	}

	return int(req.Uint32()), nil
}

// setLinkMTU uses the ioctl SIOCSIFMTU call to set the MTU of the link.
func setLinkMTU(name string, mtu uint32) error {
	req, err := unix.NewIfreq(name)
	if err != nil {
		return fmt.Errorf("ifreq: %w", err)
	}

	req.SetUint32(mtu)

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	defer unix.Close(fd)

	err = unix.IoctlIfreq(fd, unix.SIOCSIFMTU, req)
	if err != nil {
		return fmt.Errorf("ioctl: %w", err)
	}

	return nil
}
