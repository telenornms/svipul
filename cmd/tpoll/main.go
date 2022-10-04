/*
 * tpoll test program
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

// PS: While this started with a gosnmp walk-example, basically nothing is
// left of that, which is why I just removed the original copyright header,
// if you happen to be reading the git log.
package main

import (
	"os"
	"strconv"
	"fmt"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/tpoll"
	"github.com/telenornms/tpoll/omap"
	"github.com/telenornms/tpoll/session"
	"github.com/telenornms/tpoll/smierte"
)

type Task struct {
	OMap	*omap.OMap
	Mib	*smierte.Config
}

func main() {
	mib := &smierte.Config{}
	mib.Modules = []string{
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"}
	mib.Paths = []string{"mibs/std/"}
	err := mib.Init()
	if err != nil {
		tpoll.Fatalf("failed to load mibs: %s", err)
	}
	s, err := session.NewSession("192.168.122.41")
	if err != nil {
		tpoll.Fatalf("failed to start session: %s", err)
	}
	defer s.Finalize()

	t := Task{}
	t.Mib = mib
	t.OMap, err = omap.BuildOMap(s, t.Mib, "ifName")
	if err != nil {
		tpoll.Fatalf("failed to build IF-map: %s", err)
	}

	m, err := mib.Lookup(os.Args[1])
	if err != nil {
		tpoll.Fatalf("unable to lookup mib/oid/thingy: %s", err)
	}
	err = s.BulkWalk(m, t.bwCB)
	if err != nil {
		tpoll.Fatalf("Walk Failed: %v\n", err)
	}
}

func (t *Task) bwCB(pdu gosnmp.SnmpPDU) error {
	var name = pdu.Name
	if t.Mib != nil {
		n, err := t.Mib.Lookup(pdu.Name)
		if err != nil {
			tpoll.Logf("lookup failed: %s", err)
		} else {
			trailer := pdu.Name[len(n.Numeric)+1:]
			if len(trailer) > 1 {
				idxN64, _ := strconv.ParseInt(trailer[1:], 10, 32)
				idx := int(idxN64)

				if t.OMap.IdxToName[idx] != "" {
					trailer = fmt.Sprintf(".%s", t.OMap.IdxToName[idx])
				}
			}
			name = fmt.Sprintf("%s%s", n.Name, trailer)
		}
	}

	switch pdu.Type {
	case gosnmp.OctetString:
		b := pdu.Value.([]byte)
		tpoll.Logf("%s = STRING: %s\n", name, string(b))
	default:
		tpoll.Logf("%s = TYPE %d: %d\n", name, pdu.Type, gosnmp.ToBigInt(pdu.Value))
	}
	return nil
}
