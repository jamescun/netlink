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
			{16, "nlctrl", 2, 0, 0, []*Op{
				{3, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{10, CMD_CAP_DUMP | CMD_CAP_HASPOL},
			}, []*Group{
				{16, "notify"},
			}},
			{19, "netdev", 1, 0, 0, []*Op{
				{1, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{5, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{9, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{10, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{11, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{12, CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{13, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{14, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{15, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, []*Group{
				{2, "mgmt"}, {3, "page-pool"},
			}},
			{20, "ethtool", 1, 0, 0, []*Op{
				{1, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{2, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{3, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{4, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{5, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{6, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{7, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{8, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{9, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{10, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{11, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{12, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{13, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{14, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{15, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{16, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{17, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{18, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{19, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{20, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{21, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{22, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{23, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{24, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{25, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{26, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{27, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{28, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{29, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{30, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{31, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{32, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{33, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{34, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{35, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{36, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{37, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{38, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{39, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{40, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{41, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{42, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{43, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{44, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{45, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{46, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{47, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{48, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{49, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{50, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{51, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
			}, []*Group{
				{4, "monitor"},
			}},
			{21, "tcp_metrics", 1, 0, 13, []*Op{
				{1, CMD_CAP_DO | CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{2, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
			}, nil},
			{22, "nfsd", 1, 0, 0, []*Op{
				{1, CMD_CAP_DUMP | CMD_CAP_HASPOL},
				{2, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{3, CMD_CAP_DO | CMD_CAP_HASPOL},
				{4, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{5, CMD_CAP_DO | CMD_CAP_HASPOL},
				{6, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{7, CMD_CAP_DO | CMD_CAP_HASPOL},
				{8, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{9, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, nil},
			{23, "lockd", 1, 0, 0, []*Op{
				{1, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{2, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, nil},
			{24, "nbd", 1, 0, 10, []*Op{
				{1, CMD_CAP_DO | CMD_CAP_HASPOL},
				{2, CMD_CAP_DO | CMD_CAP_HASPOL},
				{3, CMD_CAP_DO | CMD_CAP_HASPOL},
				{5, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, []*Group{
				{5, "nbd_mc_group"},
			}},
			{25, "binder", 1, 0, 0, nil, []*Group{
				{6, "report"},
			}},
			{26, "SEG6", 1, 0, 7, []*Op{
				{1, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{2, ADMIN_PERM | CMD_CAP_DUMP},
				{3, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{4, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
			}, nil},
			{27, "IOAM6", 1, 0, 0, []*Op{
				{1, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{2, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{3, ADMIN_PERM | CMD_CAP_DUMP},
				{4, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{5, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{6, ADMIN_PERM | CMD_CAP_DUMP},
				{7, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
			}, []*Group{
				{7, "ioam6_events"},
			}},
			{28, "handshake", 1, 0, 0, []*Op{
				{2, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{3, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, []*Group{
				{8, "none"}, {9, "tlshd"},
			}},
			{29, "TASKSTATS", 1, 0, 0, []*Op{
				{1, ADMIN_PERM | CMD_CAP_DO | CMD_CAP_HASPOL},
				{4, CMD_CAP_DO | CMD_CAP_HASPOL},
			}, nil},
			{30, "wireguard", 1, 0, 0, []*Op{
				{0, CMD_CAP_DUMP | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
				{1, CMD_CAP_DO | CMD_CAP_HASPOL | UNS_ADMIN_PERM},
			}, nil},
		}

		target := Families{}

		for _, msg := range mr.Each {
			if msg.Header().Type == netlink.DONE || msg.Header().Type == netlink.ERROR {
				// simulate [netlink.Client] which will consume these messages
				// for us.
				break
			}

			// strip the generic netlink header.
			err := msg.UnmarshalBytes(netlink.Discard(4))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			err = target.UnmarshalNetlink(msg)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
		}

		if !cmp.Equal(expected, target) {
			t.Error(cmp.Diff(expected, target))
		}
	})
}
