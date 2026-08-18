// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package rtlink

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"go.jamescun.com/netlink"

	"golang.org/x/sys/unix"
)

// Bond is a [InfoData] for a bonded device, the parent of [BondSlave] devices.
//
// References:
//   - linux/drivers/net/bonding/bond_netlink.c
//   - https://www.kernel.org/doc/html/latest/networking/bonding.html
type Bond struct {
	Mode            BondMode
	ActiveSlave     uint32
	Miimon          time.Duration
	UpDelay         time.Duration
	DownDelay       time.Duration
	UseCarrier      bool
	ArpInterval     time.Duration
	ArpIpTarget     []netip.Addr
	ArpValidate     BondArpValidate
	ArpAllTargets   bool
	Primary         uint32
	PrimaryReselect BondPrimaryReselect
	FailOverMac     BondFailOverMac
	XmitHashPolicy  uint8
	ResendIGMP      uint32
	NumPeerNotif    uint8
	AllSlavesActive uint8
	MinLinks        uint32
	LpInterval      time.Duration
	PacketsPerSlave uint32
	AdLacpRate      uint8
	AdSelect        uint8
	AdInfo          BondAdInfo
	AdActorSysPrio  uint16
	AdUserPortKey   uint16
	AdActorSystem   net.HardwareAddr
	TlbDynamicLb    uint8
	PeerNotifyDelay time.Duration
	AdLacpActive    uint8
	MissedMax       uint8
	NsIp6Target     []netip.Addr
	CoupledControl  uint8
}

// DeviceKind returns `bond`.
func (Bond) DeviceKind() string {
	return "bond"
}

// MarshalAttributes marshals the IFLA_INFO_DATA attributes for a [Bond] device
// type.
func (Bond) MarshalAttributes(_ *netlink.AttributeEncoder) error {
	return fmt.Errorf("bond devices cannot be configured yet")
}

// UnmarshalAttributes unmarshals a [Bond] found inside [Info.Driver].
func (b *Bond) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_BOND_MODE:
			b.Mode = BondMode(attr.Uint8())
		case unix.IFLA_BOND_ACTIVE_SLAVE:
			b.ActiveSlave = attr.Uint32()
		case unix.IFLA_BOND_MIIMON:
			b.Miimon = time.Duration(attr.Uint32()) * time.Millisecond
		case unix.IFLA_BOND_UPDELAY:
			b.UpDelay = time.Duration(attr.Uint32()) * time.Millisecond
		case unix.IFLA_BOND_DOWNDELAY:
			b.DownDelay = time.Duration(attr.Uint32()) * time.Millisecond
		case unix.IFLA_BOND_USE_CARRIER:
			b.UseCarrier = true
		case unix.IFLA_BOND_ARP_INTERVAL:
			b.ArpInterval = time.Duration(attr.Uint32()) * time.Millisecond
		case unix.IFLA_BOND_ARP_IP_TARGET:
			for nested := range attr.Each {
				ip, ok := nested.Addr()
				if ok {
					b.ArpIpTarget = append(b.ArpIpTarget, ip)
				}
			}
		case unix.IFLA_BOND_ARP_VALIDATE:
			b.ArpValidate = BondArpValidate(attr.Uint32())
		case unix.IFLA_BOND_ARP_ALL_TARGETS:
			if attr.Uint32() == 1 {
				b.ArpAllTargets = true
			}
		case unix.IFLA_BOND_PRIMARY:
			b.Primary = attr.Uint32()
		case unix.IFLA_BOND_PRIMARY_RESELECT:
			b.PrimaryReselect = BondPrimaryReselect(attr.Uint32())
		case unix.IFLA_BOND_FAIL_OVER_MAC:
			b.FailOverMac = BondFailOverMac(attr.Uint8())
		case unix.IFLA_BOND_XMIT_HASH_POLICY:
			b.XmitHashPolicy = attr.Uint8()
		case unix.IFLA_BOND_RESEND_IGMP:
			b.ResendIGMP = attr.Uint32()
		case unix.IFLA_BOND_NUM_PEER_NOTIF:
			b.NumPeerNotif = attr.Uint8()
		case unix.IFLA_BOND_ALL_SLAVES_ACTIVE:
			b.AllSlavesActive = attr.Uint8()
		case unix.IFLA_BOND_MIN_LINKS:
			b.MinLinks = attr.Uint32()
		case unix.IFLA_BOND_LP_INTERVAL:
			b.LpInterval = time.Duration(attr.Uint32()) * time.Second
		case unix.IFLA_BOND_PACKETS_PER_SLAVE:
			b.PacketsPerSlave = attr.Uint32()
		case unix.IFLA_BOND_AD_LACP_RATE:
			b.AdLacpRate = attr.Uint8()
		case unix.IFLA_BOND_AD_SELECT:
			b.AdSelect = attr.Uint8()
		case unix.IFLA_BOND_AD_INFO:
			err := attr.Unmarshal(&b.AdInfo)
			if err != nil {
				return err
			}
		case unix.IFLA_BOND_AD_ACTOR_SYS_PRIO:
			b.AdActorSysPrio = attr.Uint16()
		case unix.IFLA_BOND_AD_USER_PORT_KEY:
			b.AdUserPortKey = attr.Uint16()
		case unix.IFLA_BOND_AD_ACTOR_SYSTEM:
			b.AdActorSystem = attr.Bytes()
		case unix.IFLA_BOND_TLB_DYNAMIC_LB:
			b.TlbDynamicLb = attr.Uint8()
		case unix.IFLA_BOND_PEER_NOTIF_DELAY:
			b.PeerNotifyDelay = time.Duration(attr.Uint32()) * time.Millisecond
		case unix.IFLA_BOND_AD_LACP_ACTIVE:
			b.AdLacpActive = attr.Uint8()
		case unix.IFLA_BOND_MISSED_MAX:
			b.MissedMax = attr.Uint8()
		case unix.IFLA_BOND_NS_IP6_TARGET:
			for item := range attr.Each {
				ip, ok := netip.AddrFromSlice(item.Bytes())
				if ok {
					b.NsIp6Target = append(b.NsIp6Target, ip)
				}
			}
		case unix.IFLA_BOND_COUPLED_CONTROL:
			b.CoupledControl = attr.Uint8()
		}
	}

	return nil
}

// BondAdInfo contains information about a [Bond] device 802.3ad bonding
// configuration.
type BondAdInfo struct {
	Aggregator uint16
	NumPorts   uint16
	ActorKey   uint16
	PartnerKey uint16
	PartnerMac net.HardwareAddr
}

// UnmarshalAttributes unmarshals the [BondAdInfo] sub-message for [Bond].
func (b *BondAdInfo) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_BOND_AD_INFO_AGGREGATOR:
			b.Aggregator = attr.Uint16()
		case unix.IFLA_BOND_AD_INFO_NUM_PORTS:
			b.NumPorts = attr.Uint16()
		case unix.IFLA_BOND_AD_INFO_ACTOR_KEY:
			b.ActorKey = attr.Uint16()
		case unix.IFLA_BOND_AD_INFO_PARTNER_KEY:
			b.PartnerKey = attr.Uint16()
		case unix.IFLA_BOND_AD_INFO_PARTNER_MAC:
			b.PartnerMac = attr.Bytes()
		}
	}

	return nil
}

// BondFailOverMac configures the MAC address selecting during a failover event
// on a [Bond] device.
type BondFailOverMac uint8

// Constants for [BondFailOverMac].
const (
	BOND_MAC_ACTIVE BondFailOverMac = 1 + iota
	BOND_MAC_FOLLOW
)

func (b BondFailOverMac) String() string {
	switch b {
	case 0:
		return "NONE"
	case BOND_MAC_ACTIVE:
		return "ACTIVE"
	case BOND_MAC_FOLLOW:
		return "FOLLOW"

	default:
		return "UNKNOWN"
	}
}

// BondArpValidate configures how a [Bond] device validates ARP probes.
type BondArpValidate uint32

// Constants for [BondArpValidate].
const (
	BOND_ACTIVE BondArpValidate = 1 + iota
	BOND_BACKUP
	BOND_ALL
	BOND_FILTER
	BOND_FILTER_ACTIVE
	BOND_FILTER_BACKUP
)

func (b BondArpValidate) String() string {
	switch b {
	case 0:
		return "NONE"
	case BOND_ACTIVE:
		return "ACTIVE"
	case BOND_BACKUP:
		return "BACKUP"
	case BOND_ALL:
		return "ALL"
	case BOND_FILTER:
		return "FILTER"
	case BOND_FILTER_ACTIVE:
		return "FILTER_ACTIVE"
	case BOND_FILTER_BACKUP:
		return "FILTER_BACKUP"

	default:
		return "UNKNOWN"
	}
}

// BondMode configures the bonding policy of a [Bond] device.
type BondMode uint8

// Constants for [BondMode].
const (
	BOND_BALANCE_RR BondMode = iota
	BOND_ACTIVE_BACKUP
	BOND_BALANCE_XOR
	BOND_BROADCAST
	BOND_8023AD
	BOND_BALANCE_TLB
	BOND_BALANCE_ALB
)

func (b BondMode) String() string {
	switch b {
	case BOND_BALANCE_RR:
		return "BALANCE-RR"
	case BOND_ACTIVE_BACKUP:
		return "ACTIVE-BACKUP"
	case BOND_BALANCE_XOR:
		return "BALANCE_XOR"
	case BOND_BROADCAST:
		return "BROADCAST"
	case BOND_8023AD:
		return "802.3AD"
	case BOND_BALANCE_TLB:
		return "BALANCE_TLB"
	case BOND_BALANCE_ALB:
		return "BALANCE_ALB"

	default:
		return "UNKNOWN"
	}
}

// BondPrimaryReselect configured the reselection policy of a [Bond] device.
type BondPrimaryReselect uint32

// Constants for [BondPrimaryReselect].
const (
	BOND_ALWAYS BondPrimaryReselect = iota
	BOND_BETTER
	BOND_FAILURE
)

func (b BondPrimaryReselect) String() string {
	switch b {
	case BOND_ALWAYS:
		return "ALWAYS"
	case BOND_BETTER:
		return "BETTER"
	case BOND_FAILURE:
		return "FAILURE"

	default:
		return "UNKNOWN"
	}
}

// BondSlave is a [Bond] slave device.
//
// References:
//   - linux/drivers/net/bonding/bond_netlink.c
type BondSlave struct {
	State                  uint8
	Miistatus              uint8
	LinkFailureCount       uint32
	PermAddress            net.HardwareAddr
	QueueId                uint16
	AdAggregatorId         uint16
	AdActorOperPortState   uint8
	AdPartnerOperPortState uint8
	Prio                   int32
}

// SlaveKind returns `bond`.
func (BondSlave) SlaveKind() string {
	return "bond"
}

// UnmarshalAttributes unmarshals a [Bond] device slave.
func (b *BondSlave) UnmarshalAttributes(attrs *netlink.AttributeDecoder) error {
	for attr := range attrs.Each {
		switch attr.Type() {
		case unix.IFLA_BOND_SLAVE_STATE:
			b.State = attr.Uint8()
		case unix.IFLA_BOND_SLAVE_MII_STATUS:
			b.Miistatus = attr.Uint8()
		case unix.IFLA_BOND_SLAVE_LINK_FAILURE_COUNT:
			b.LinkFailureCount = attr.Uint32()
		case unix.IFLA_BOND_SLAVE_PERM_HWADDR:
			b.PermAddress = attr.Bytes()
		case unix.IFLA_BOND_SLAVE_QUEUE_ID:
			b.QueueId = attr.Uint16()
		case unix.IFLA_BOND_SLAVE_AD_AGGREGATOR_ID:
			b.AdAggregatorId = attr.Uint16()
		case unix.IFLA_BOND_SLAVE_AD_ACTOR_OPER_PORT_STATE:
			b.AdActorOperPortState = attr.Uint8()
		case unix.IFLA_BOND_SLAVE_AD_PARTNER_OPER_PORT_STATE:
			b.AdPartnerOperPortState = attr.Uint8()
		case unix.IFLA_BOND_SLAVE_PRIO:
			b.Prio = attr.Int32()
		}
	}

	return nil
}
