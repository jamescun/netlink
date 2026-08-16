// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// Package known contains a description of known Generic Netlink families,
// their commands and attributes, used for debugging output.
package known

// Family contains a description of a known Netlink Family.
type Family struct {
	Name string
	Desc string
	Cmds []Cmd
}

// GetFamily retrieves a known Generic Netlink family by name.
func GetFamily(name string) (family Family, found bool) {
	for _, family = range families {
		if family.Name == name {
			found = true
			return
		}
	}

	return
}

// GetCmd retrieves a known Generic Netlink command/operation from a family.
func (f Family) GetCmd(id uint8) (cmd Cmd, found bool) {
	for _, cmd = range f.Cmds {
		if cmd.Cmd == id {
			found = true
			return
		}
	}

	return
}

// Cmd contains a description of a command of a known family.
type Cmd struct {
	Cmd  uint8
	Name string
	Desc string
}
