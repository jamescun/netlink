// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package genetlink

import (
	"os"
	"testing"

	"go.jamescun.com/netlink"

	"github.com/google/go-cmp/cmp"
)

func TestFamilies(t *testing.T) {
	t.Run("UnmarshalNetlink", func(t *testing.T) {
		bytes, err := os.ReadFile("testdata/list_families")
		if err != nil {
			t.Fatal("could not load testdata:", err)
		}

		mr := netlink.NewMessageReader(bytes)

		expected := Families{
			{16, "nlctrl", 2, 0, 0, []McastGroup{
				{16, "notify"},
			}},
			{19, "netdev", 1, 0, 0, []McastGroup{
				{2, "mgmt"}, {3, "page-pool"},
			}},
			{20, "ethtool", 1, 0, 0, []McastGroup{
				{4, "monitor"},
			}},
			{21, "tcp_metrics", 1, 0, 13, nil},
			{22, "nfsd", 1, 0, 0, nil},
			{23, "lockd", 1, 0, 0, nil},
			{24, "nbd", 1, 0, 10, []McastGroup{
				{5, "nbd_mc_group"},
			}},
			{25, "binder", 1, 0, 0, []McastGroup{
				{6, "report"},
			}},
			{26, "SEG6", 1, 0, 7, nil},
			{27, "IOAM6", 1, 0, 0, []McastGroup{
				{7, "ioam6_events"},
			}},
			{28, "handshake", 1, 0, 0, []McastGroup{
				{8, "none"}, {9, "tlshd"},
			}},
			{29, "TASKSTATS", 1, 0, 0, nil},
			{30, "wireguard", 1, 0, 0, nil},
		}

		target := Families{}

		for _, msg := range mr.Each {
			err := target.UnmarshalNetlink(msg)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
		}

		if !cmp.Equal(expected, target) {
			t.Error(cmp.Diff(expected, target))
		}
	})
}
