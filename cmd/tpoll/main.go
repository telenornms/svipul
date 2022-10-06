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

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/tpoll"
	"github.com/telenornms/tpoll/omap"
	"github.com/telenornms/tpoll/session"
	"github.com/telenornms/tpoll/smierte"
)

type Task struct {
	OMap   *omap.OMap
	Mib    *smierte.Config
	Metric skogul.Metric
}

func main() {
	mib := &smierte.Config{}
	mib.Modules = []string{
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"}
	mib.Paths = []string{"mibs/modules/"}
	err := mib.Init()
	if err != nil {
		tpoll.Fatalf("failed to load mibs: %s", err)
	}
	config, err := config.Path("skogul")
	if err != nil {
		tpoll.Fatalf("Failed to configure Skogul: %v", err)
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

	m := make([]tpoll.Node, 0, len(os.Args)-1)
	for idx, arg := range os.Args {
		if idx == 0 {
			continue
		}
		nym, err := mib.Lookup(arg)
		if err != nil {
			tpoll.Fatalf("unable to lookup mib/oid/thingy %s: %s", arg, err)
		}
		m = append(m, nym)
	}
	if len(m) < 1 {
		tpoll.Fatalf("no oids to look up?")
	}
	t.Metric.Metadata = make(map[string]interface{})
	t.Metric.Data = make(map[string]interface{})
	err = s.BulkWalk(m, t.bwCB)
	if err != nil {
		tpoll.Fatalf("Walk Failed: %v\n", err)
	}
	c := skogul.Container{}
	c.Metrics = append(c.Metrics, &t.Metric)

	err = config.Handlers["tpoll"].Handler.TransformAndSend(&c)
	if err != nil {
		tpoll.Fatalf("sending failed: %v", err)
	}
}

func (t *Task) bwCB(pdu gosnmp.SnmpPDU) error {
	var name = pdu.Name
	var element = ""
	if t.Mib != nil {
		n, err := t.Mib.Lookup(pdu.Name)
		if err != nil {
			tpoll.Logf("lookup failed: %s", err)
		} else {
			trailer := pdu.Name[len(n.Numeric)+1:][1:]
			if len(trailer) > 0 {
				idxN64, _ := strconv.ParseInt(trailer, 10, 32)
				idx := int(idxN64)

				if t.OMap.IdxToName[idx] != "" {
					trailer = t.OMap.IdxToName[idx]
				}
			}
			name = n.Name
			element = trailer
		}
	}

	if t.Metric.Data[element] == nil {
		t.Metric.Data[element] = make(map[string]interface{})
	}
	switch pdu.Type {
	case gosnmp.OctetString:
		b := pdu.Value.([]byte)
		(t.Metric.Data[element].(map[string]interface{}))[name] = string(b)
	default:
		(t.Metric.Data[element].(map[string]interface{}))[name] = gosnmp.ToBigInt(pdu.Value)
	}
	return nil
}
