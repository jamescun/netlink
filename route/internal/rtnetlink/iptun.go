// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtnetlink

// Constants for IPIP tunnel attributes.
const (
	IPTUN_UNSPEC = iota
	IPTUN_LINK
	IPTUN_LOCAL
	IPTUN_REMOTE
	IPTUN_TTL
	IPTUN_TOS
	IPTUN_ENCAP_LIMIT
	IPTUN_FLOWINFO
	IPTUN_FLAGS
	IPTUN_PROTO
	IPTUN_PMTUDISC
	IPTUN_6RD_PREFIX
	IPTUN_6RD_RELAY_PREFIX
	IPTUN_6RD_PREFIXLEN
	IPTUN_6RD_RELAY_PREFIXLEN
	IPTUN_ENCAP_TYPE
	IPTUN_ENCAP_FLAGS
	IPTUN_ENCAP_SPORT
	IPTUN_ENCAP_DPORT
	IPTUN_COLLECT_METADATA
	IPTUN_FWMARK
)
