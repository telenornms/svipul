/*
 * svpul config
 *
 * Copyright (c) 2022 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
 * 02110-1301  USA
 */

package svipul

import (
	"time"
)

type conf struct {
	DefaultCommunity string
	Workers          int
	Debug            bool
	MibPaths         []string
	MibModules       []string
	MaxMapAge        time.Duration
}

var Config conf = conf{
	DefaultCommunity: "public",
	Debug:            false,
	MibPaths:         []string{"mibs/modules"},
	MaxMapAge:        time.Second * 3600,
	MibModules: []string{
		"ADSL-LINE-MIB",
		"ADSL-TC-MIB",
		"ALARM-MIB",
		"ATM-FORUM-MIB",
		"ATM-MIB",
		"ATM-TC-MIB",
		"BGP4-MIB",
		"BRIDGE-MIB",
		"DIFFSERV-DSCP-TC",
		"DIFFSERV-MIB",
		"DISMAN-PING-MIB",
		"DISMAN-TRACEROUTE-MIB",
		"DLSW-MIB",
		"DRAFT-MSDP-MIB",
		"DS1-MIB",
		"DS3-MIB",
		"ENTITY-MIB",
		"ENTITY-STATE-MIB",
		"ENTITY-STATE-TC-MIB",
		"ESO-CONSORTIUM-MIB",
		"EtherLike-MIB",
		"ETHER-WIS",
		"FRAME-RELAY-DTE-MIB",
		"FR-MFR-MIB",
		"GGSN-MIB",
		"GMPLS-LSR-STD-MIB",
		"GMPLS-TC-STD-MIB",
		"GMPLS-TE-STD-MIB",
		"HCNUM-TC",
		"HOST-RESOURCES-MIB",
		"HOST-RESOURCES-TYPES",
		"IANA-ADDRESS-FAMILY-NUMBERS-MIB",
		"IANA-GMPLS-TC-MIB",
		"IANAifType-MIB",
		"IANA-RTPROTO-MIB",
		"IEEE8021-CFM-MIB",
		"IEEE8021-CFM-V2-MIB",
		"IEEE8021-PAE-MIB",
		"IEEE8021-TC-MIB",
		"IEEE8023-LAG-MIB",
		"IF-MIB",
		"IGMP-STD-MIB",
		"INET-ADDRESS-MIB",
		"INTEGRATED-SERVICES-MIB",
		"IP-FORWARD-MIB",
		"IP-MIB",
		"IPMROUTE-MIB",
		"IPMROUTE-STD-MIB",
		"IPV6-FLOW-LABEL-MIB",
		"IPV6-MIB",
		"IPV6-TC",
		"ISIS-MIB",
		"LLDP-MIB",
		"MPLS-L3VPN-STD-MIB",
		"MPLS-LSR-STD-MIB",
		"MPLS-TC-STD-MIB",
		"MPLS-TE-STD-MIB",
		"MPLS-VPN-MIB",
		"OPT-IF-MIB",
		"OSPF-MIB",
		"OSPF-TRAP-MIB",
		"OSPFV3-MIB",
		"P-BRIDGE-MIB",
		"PerfHist-TC-MIB",
		"PIM-MIB",
		"POWER-ETHERNET-MIB",
		"PPP-LCP-MIB",
		"PTOPO-MIB",
		"Q-BRIDGE-MIB",
		"RADIUS-ACC-CLIENT-MIB",
		"RADIUS-AUTH-CLIENT-MIB",
		"RMON2-MIB",
		"RMON-MIB",
		"RSTP-MIB",
		"SNA-SDLC-MIB",
		"SNMP-COMMUNITY-MIB",
		"SNMP-FRAMEWORK-MIB",
		"SNMP-MPD-MIB",
		"SNMP-NOTIFICATION-MIB",
		"SNMP-PROXY-MIB",
		"SNMP-TARGET-MIB",
		"SNMP-USER-BASED-SM-MIB",
		"SNMPv2-MIB",
		"SNMPv2-SMI",
		"SNMPv2-TC",
		"SNMPv2-TM",
		"SNMP-VIEW-BASED-ACM-MIB",
		"SONET-MIB",
		"SYSAPPL-MIB",
		"TCP-MIB",
		"TOKEN-RING-RMON-MIB",
		"TRANSPORT-ADDRESS-MIB",
		"TUNNEL-MIB",
		"UDP-MIB",
		"VPN-TC-STD-MIB",
		"VRRP-MIB",
		"VRRPV3-MIB"},
}
