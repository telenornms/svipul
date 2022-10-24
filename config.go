/*
 * tpoll config
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

package tpoll

import (
	"time"
	)

type conf struct {
	DefaultCommunity string
	Workers          int
	Debug            bool
	MibPaths         []string
	MibModules       []string
	MaxMapAge	 time.Duration
}

var Config conf = conf{
	DefaultCommunity: "public",
	Debug:            true,
	MibPaths:         []string{"mibs/modules"},
	MaxMapAge:	time.Second*60,
	MibModules: []string{
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"},
}
