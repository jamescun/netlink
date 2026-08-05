// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// nlctrl is a utility that lists the available Generic Netlink families on
// the system.
package main

import (
	"fmt"
	"os"

	"go.jamescun.com/netlink"
	"go.jamescun.com/netlink/genetlink"
)

func main() {
	client, err := netlink.Connect(netlink.GENERIC)
	if err != nil {
		exitError(1, "could not connect to netlink: %s", err)
	}
	defer client.Close()

	ctrl, err := genetlink.NewController(client)
	if err != nil {
		exitError(1, "could not connect to genetlink controller: %s", err)
	}

	families, err := ctrl.ListFamilies()
	if err != nil {
		exitError(1, "could not list genetlink families: %s", err)
	}

	fmt.Println("ID   NAME               VERSION   MULTICAST GROUPS")
	fmt.Println("--   ----               -------   ----------------")

	for _, family := range families {
		fmt.Printf("%-4d %-18s %-9d", family.ID, family.Name, family.Version)

		for _, grp := range family.McastGroups {
			fmt.Printf(" 0x%02d:%s", grp.ID, grp.Name)
		}

		fmt.Println()
	}
}

// exitError writes an error message to stderr and exits the given exit code.
func exitError(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(code)
}
