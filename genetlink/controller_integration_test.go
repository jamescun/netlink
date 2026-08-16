// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

//go:build integration

package genetlink_test

import (
	"errors"
	"testing"

	"go.jamescun.com/netlink/genetlink"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sys/unix"
)

func TestController(t *testing.T) {
	ctrl, err := genetlink.NewController()
	if err != nil {
		t.Fatal("could not create controller:", err)
	}
	defer ctrl.Close()

	t.Run("GetFamily", func(t *testing.T) {
		expected := &genetlink.Family{
			ID:      unix.GENL_ID_CTRL,
			Name:    "nlctrl",
			Version: 2,
			Ops: []*genetlink.Op{
				{3, genetlink.CMD_CAP_DO | genetlink.CMD_CAP_DUMP | genetlink.CMD_CAP_HASPOL},
				{10, genetlink.CMD_CAP_DUMP | genetlink.CMD_CAP_HASPOL},
			},
			Groups: []*genetlink.Group{
				{ID: 16, Name: "notify"},
			},
		}

		target, err := ctrl.GetFamily("nlctrl")
		if err != nil {
			t.Fatal("unexpected error:", err)
		}

		if !cmp.Equal(expected, target) {
			t.Error(cmp.Diff(expected, target))
		}
	})

	t.Run("GetFamily/NotFound", func(t *testing.T) {
		_, err := ctrl.GetFamily("not_found")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, genetlink.ErrFamilyNotFound) {
			t.Errorf("expected error:\n%v\ngot:\n%v", genetlink.ErrFamilyNotFound, err)
		}
	})
}
