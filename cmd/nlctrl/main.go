// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

// nlctrl is a utility to get information about the registered Generic Netlink
// families on the system.
package main

import (
	"fmt"
	"io"
	"os"

	"go.jamescun.com/netlink/genetlink"
	"go.jamescun.com/netlink/genetlink/known"
)

// version and commit information embedded through linker flags.
var (
	version = `0.0.0`
	commit  = `unknown`
)

// usage information printed when -h/--help is requested.
const usage = `nlctrl v%s %s
get generic netlink family information

USAGE: nlctrl [options...] [family name]

FAMILY NAME:
  When a family name is specified, only information about that family is
  returned. Otherwise all families registered on the system will be returned.

OPTIONS:
  -h --help  show usage information
`

func main() {
	w := os.Stdout
	args := os.Args[1:]

	ctrl, err := genetlink.NewController()
	if err != nil {
		exitError(w, err)
	}
	defer ctrl.Close()

	switch len(args) {
	case 0:
		// list all families.
		families, err := ctrl.ListFamilies()
		if err != nil {
			exitError(w, err)
		}

		for _, family := range families {
			exitError(w, printFamily(w, family))
		}

	case 1:
		// get a specific family.
		if args[0] == "-h" || args[0] == "--help" {
			exitError(w, printUsage(w))
			break
		}

		family, err := ctrl.GetFamily(args[0])
		if err != nil {
			exitError(w, err)
		}

		exitError(w, printFamily(w, family))

	default:
		exitError(w, printUsage(w))
	}
}

// printFamily prints information about a single Generic Netlink family.
func printFamily(w io.Writer, family *genetlink.Family) error {
	const (
		tplFamily  = "Family:   %d\nName:     %s v%d\n"
		tplDesc    = "Desc:     %s\n"
		tplHdrSize = "HdrSize:  %d\n"
		tplMaxAttr = "MaxAttr:  %d\n"
		tplCmd     = "  %v\n"
		tplFlags   = "    Flags:  %s\n"
		tplGroup   = "    %2d: %s\n"
	)

	info, found := known.GetFamily(family.Name)

	_, err := fmt.Fprintf(w, tplFamily, family.ID, family.Name, family.Version)
	if err != nil {
		return err
	}

	if found {
		if info.Desc != "" {
			_, err = fmt.Fprintf(w, tplDesc, info.Desc)
			if err != nil {
				return err
			}
		}
	}

	if family.HdrSize != 0 {
		_, err = fmt.Fprintf(w, tplHdrSize, family.HdrSize)
		if err != nil {
			return err
		}
	}

	if family.MaxAttr > 0 {
		_, err = fmt.Fprintf(w, tplMaxAttr, family.MaxAttr)
		if err != nil {
			return err
		}
	}

	if len(family.Ops) > 0 {
		_, err = fmt.Fprint(w, "\nCommands:\n")
		if err != nil {
			return err
		}

		for _, op := range family.Ops {
			cmd, found := info.GetCmd(uint8(op.ID)) //nolint

			if found {
				_, err = fmt.Fprintf(w, tplCmd, cmd.Name)
				if err != nil {
					return err
				}
			} else {
				_, err = fmt.Fprintf(w, tplCmd, op.ID)
				if err != nil {
					return err
				}
			}

			_, err = fmt.Fprintf(w, tplFlags, op.Flags)
			if err != nil {
				return err
			}
		}
	}

	if len(family.Groups) > 0 {
		_, err = fmt.Fprint(w, "\nMulticast Groups:\n")
		if err != nil {
			return err
		}

		for _, group := range family.Groups {
			_, err = fmt.Fprintf(w, tplGroup, group.ID, group.Name)
			if err != nil {
				return err
			}
		}
	}

	_, err = fmt.Fprint(w, "\n")
	if err != nil {
		return err
	}

	return err
}

// printUsage prints the nlctrl usage information.
func printUsage(w io.Writer) error {
	commit := commit
	if len(commit) > 7 {
		// only show a truncated commit hash.
		commit = commit[:7]
	}

	_, err := fmt.Fprintf(w, usage, version, commit)
	return err
}

func exitError(w io.Writer, err error) {
	if err != nil {
		fmt.Fprintln(w, "error:", err) //nolint
		os.Exit(1)
	}
}
