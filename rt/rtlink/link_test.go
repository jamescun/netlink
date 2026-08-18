// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"net"
	"net/netip"
	"os"
	"testing"

	"go.jamescun.com/netlink"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/sys/unix"
)

func TestLink(t *testing.T) {
	t.Run("UnmarshalNetlink", func(t *testing.T) {
		tests := []struct {
			path     string
			expected *Link
		}{
			{
				"testdata/getlink_bridge",
				&Link{
					Type:      EtherType(unix.ARPHRD_ETHER),
					Index:     17,
					Flags:     BROADCAST | MULTICAST,
					Address:   net.HardwareAddr{0x3e, 0x05, 0xdd, 0x8d, 0x58, 0x21},
					Broadcast: net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
					Name:      "br0",
					MTU:       1500,
					Qdisc:     "noop",
					Stats:     &Stats{},
					TxQlen:    1000,
					State:     STATE_DOWN,
					Info: Info{
						Kind: "bridge",
						Data: &Bridge{
							ForwardDelay:            1500,
							HelloTime:               200,
							MaxAge:                  2000,
							AgeingTime:              30000,
							Priority:                32768,
							VlanProtocol:            129,
							RootID:                  BridgeID{Prio: 128, Addr: net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
							BridgeID:                BridgeID{Prio: 128, Addr: net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
							GroupAddr:               net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x00},
							McastRouter:             1,
							McastSnooping:           true,
							McastHashElasticity:     16,
							McastHashMax:            4096,
							McastLastMemberCnt:      2,
							McastStartupQueryCnt:    2,
							McastLastMemberIntvl:    100,
							McastMembershipIntvl:    26000,
							McastQuerierIntvl:       25500,
							McastQueryIntvl:         12500,
							McastQueryResponseIntvl: 1000,
							McastStartupQueryIntvl:  3125,
							VlanDefaultPvid:         1,
							McastIgmpVersion:        2,
							McastMldVersion:         1,
						},
					},
					Spec: Spec{
						Inet: &Inet{
							Config: InetConfig{
								Forwarding:         true,
								AcceptRedirects:    true,
								SecureRedirects:    true,
								SendRedirects:      true,
								SharedMedia:        true,
								RpFilter:           STRICT,
								PromoteSecondaries: true,
							},
						},
						Inet6: &Inet6{
							Config: Inet6Config{
								HopLimit:                       64,
								Mtu:                            1500,
								AcceptRa:                       ACCEPT,
								AcceptRedirects:                true,
								AutoConf:                       true,
								DadTransmits:                   1,
								RtrSolicits:                    4294967295,
								RtrSolicitInterval:             4000000000000,
								RtrSolicitMaxInterval:          1000000000000,
								ForceMldVersion:                172800,
								Mldv1UnsolicitedReportInterval: 86400000000,
								Mldv2UnsolicitedReportInterval: 3000000,
								UseTempAddr:                    600,
								TempValidLft:                   16,
								RegenMaxRetry:                  1,
								MaxDesyncFactor:                1,
								AcceptRaRtInfoMaxPlen:          1,
								OptimisticDad:                  true,
								UseOptimistic:                  true,
								McForwarding:                   1,
								ForceTllao:                     true,
								AcceptRaFromLocal:              true,
								DropUnsolicitedNa:              true,
								Seg6RequireHmac:                1,
								Ioam6Enabled:                   1024,
								Ioam6IdWide:                    65535,
								NdiscEvictNoCarrier:            true,
								AcceptUntrackedNa:              1,
							},
							CacheInfo: Inet6CacheInfo{
								MaxReasmLen:   65535,
								Tstamp:        44036906,
								ReachableTime: 38391,
								RetransTime:   1000,
							},
							Token: netip.MustParseAddr("::"),
						},
					},
					NumTxQueues:    1,
					NumRxQueues:    1,
					Carrier:        1,
					GsoMaxSegs:     65535,
					GsoMaxSize:     65536,
					MinMTU:         68,
					MaxMTU:         65535,
					GroMaxSize:     65536,
					TsoMaxSize:     65536,
					TsoMaxSegs:     65535,
					GsoIpv4MaxSize: 65536,
					GroIpv4MaxSize: 65536,
				},
			},
			{
				"testdata/getlink_veth",
				&Link{
					Type:      EtherType(unix.ARPHRD_ETHER),
					Index:     2,
					Flags:     UP | BROADCAST | RUNNING | MULTICAST | LOWER_UP,
					Address:   net.HardwareAddr{0xf6, 0xbe, 0x74, 0x08, 0x82, 0x1c},
					Broadcast: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					Name:      "eth0",
					MTU:       1500,
					Link:      7,
					Qdisc:     "noqueue",
					Stats:     &Stats{RxPackets: 6094, TxPackets: 2000, RxBytes: 16602799, TxBytes: 294591},
					TxQlen:    1000,
					State:     STATE_UP,
					Info: Info{
						Kind: "veth",
					},
					Spec: Spec{
						Inet: &Inet{
							Config: InetConfig{
								Forwarding:         true,
								AcceptRedirects:    true,
								SecureRedirects:    true,
								SendRedirects:      true,
								SharedMedia:        true,
								RpFilter:           STRICT,
								PromoteSecondaries: true,
							},
						},
						Inet6: &Inet6{
							Flags: READY,
							Config: Inet6Config{
								HopLimit:                       64,
								Mtu:                            1500,
								AcceptRedirects:                true,
								AutoConf:                       true,
								DadTransmits:                   1,
								RtrSolicits:                    4294967295,
								RtrSolicitInterval:             4000000000000,
								RtrSolicitMaxInterval:          1000000000000,
								ForceMldVersion:                172800,
								Mldv1UnsolicitedReportInterval: 86400000000,
								Mldv2UnsolicitedReportInterval: 3000000,
								UseTempAddr:                    600,
								TempValidLft:                   16,
								RegenMaxRetry:                  1,
								MaxDesyncFactor:                1,
								OptimisticDad:                  true,
								UseOptimistic:                  true,
								McForwarding:                   1,
								ForceTllao:                     true,
								AcceptRaFromLocal:              true,
								DropUnsolicitedNa:              true,
								Seg6RequireHmac:                1,
								EnhancedDad:                    true,
								Ioam6Enabled:                   1024,
								Ioam6IdWide:                    65535,
								NdiscEvictNoCarrier:            true,
								AcceptUntrackedNa:              1,
							},
							CacheInfo: Inet6CacheInfo{
								MaxReasmLen:   65535,
								Tstamp:        75,
								ReachableTime: 21466,
								RetransTime:   1000,
							},
							Token:       netip.MustParseAddr("::"),
							AddrGenMode: 1,
						},
					},
					NumTxQueues:      8,
					NumRxQueues:      8,
					Carrier:          1,
					CarrierChanges:   4,
					GsoMaxSegs:       65535,
					GsoMaxSize:       65536,
					CarrierUpCount:   2,
					CarrierDownCount: 2,
					MinMTU:           68,
					MaxMTU:           65535,
					GroMaxSize:       65536,
					TsoMaxSize:       524280,
					TsoMaxSegs:       65535,
					GsoIpv4MaxSize:   65536,
					GroIpv4MaxSize:   65536,
				},
			},
			{
				"testdata/getlink_vlan",
				&Link{
					Type:      EtherType(unix.ARPHRD_ETHER),
					Index:     16,
					Flags:     UP | BROADCAST | RUNNING | MULTICAST | LOWER_UP,
					Address:   net.HardwareAddr{0xf6, 0xbe, 0x74, 0x08, 0x82, 0x1c},
					Broadcast: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					Name:      "eth0.10",
					MTU:       1500,
					Link:      2,
					Qdisc:     "noqueue",
					Stats:     &Stats{RxPackets: 10, TxPackets: 10, RxBytes: 656, TxBytes: 796, Multicast: 10},
					TxQlen:    1000,
					State:     STATE_UP,
					Info: Info{
						Kind: "vlan",
						Data: &VLAN{
							ID:       10,
							Flags:    REORDER_HDR,
							Protocol: 129,
						},
					},
					Spec: Spec{
						Inet: &Inet{
							Config: InetConfig{
								Forwarding:         true,
								AcceptRedirects:    true,
								SecureRedirects:    true,
								SendRedirects:      true,
								SharedMedia:        true,
								RpFilter:           STRICT,
								PromoteSecondaries: true,
							},
						},
						Inet6: &Inet6{
							Flags: RS_SENT | READY,
							Config: Inet6Config{
								HopLimit:                       64,
								Mtu:                            1500,
								AcceptRa:                       ACCEPT,
								AcceptRedirects:                true,
								AutoConf:                       true,
								DadTransmits:                   1,
								RtrSolicits:                    4294967295,
								RtrSolicitInterval:             4000000000000,
								RtrSolicitMaxInterval:          1000000000000,
								ForceMldVersion:                172800,
								Mldv1UnsolicitedReportInterval: 86400000000,
								Mldv2UnsolicitedReportInterval: 3000000,
								UseTempAddr:                    600,
								TempValidLft:                   16,
								RegenMaxRetry:                  1,
								MaxDesyncFactor:                1,
								AcceptRaRtInfoMaxPlen:          1,
								OptimisticDad:                  true,
								UseOptimistic:                  true,
								McForwarding:                   1,
								ForceTllao:                     true,
								AcceptRaFromLocal:              true,
								DropUnsolicitedNa:              true,
								Seg6RequireHmac:                1,
								Ioam6Enabled:                   1024,
								Ioam6IdWide:                    65535,
								NdiscEvictNoCarrier:            true,
								AcceptUntrackedNa:              1,
							},
							CacheInfo: Inet6CacheInfo{
								MaxReasmLen:   65535,
								Tstamp:        43293365,
								ReachableTime: 38981,
								RetransTime:   1000,
							},
							Token: netip.MustParseAddr("::"),
						},
					},
					NumTxQueues:    1,
					NumRxQueues:    1,
					Carrier:        1,
					GsoMaxSegs:     65535,
					GsoMaxSize:     65536,
					MaxMTU:         65535,
					GroMaxSize:     65536,
					TsoMaxSize:     524280,
					TsoMaxSegs:     65535,
					GsoIpv4MaxSize: 65536,
					GroIpv4MaxSize: 65536,
				},
			},
		}

		for _, test := range tests {
			t.Run(test.path, func(t *testing.T) {
				opts := []cmp.Option{
					cmpopts.EquateComparable(netip.Addr{}),
				}

				bytes, err := os.ReadFile(test.path)
				if err != nil {
					t.Fatal("could not load testdata:", err)
				}

				msg, err := netlink.NewMessageDecoder(bytes)
				if err != nil {
					t.Fatal("could not unmarshal test message:", err)
				}

				target := &Link{}

				err = target.UnmarshalNetlink(msg)
				if test.expected == nil {
					if err == nil {
						t.Fatal("expected error")
					}
				} else {
					if err != nil {
						t.Fatal("unexpected error:", err)
					}

					if !cmp.Equal(test.expected, target, opts...) {
						t.Error(cmp.Diff(test.expected, target, opts...))
					}
				}
			})
		}
	})
}
