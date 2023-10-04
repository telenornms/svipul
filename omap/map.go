/*
 * tpoll map
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

package omap

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/svipul"
	"github.com/telenornms/svipul/smierte"
	"time"
)

// OMap is a two-way map of index to name, the typical case is ifIndex to
// ifName, but can be anything.
type OMap struct {
	IdxToName map[string]string
	NameToIdx map[string]string
	Oid       tpoll.Node // OID used to build the map, e.g.: ifName
	Timestamp time.Time  // When was the map created?
}

func BuildOMap(w tpoll.Walker, oid string) (*OMap, error) {
	m := &OMap{}
	var err error
	m.IdxToName = make(map[string]string)
	m.NameToIdx = make(map[string]string)
	m.Timestamp = time.Now()
	m.Oid, err = smierte.Lookup(oid)
	if err != nil {
		return nil, fmt.Errorf("lookup of oid %s failed: %w", oid, err)
	}
	if m.Oid.Numeric == "" {
		return nil, fmt.Errorf("what happened with mib.Lookup? m.Oid: %#v", m.Oid)
	}
	err = w.BulkWalk([]tpoll.Node{m.Oid}, m.walkCB)
	since := time.Since(m.Timestamp).Round(time.Millisecond * 100)
	if err == nil {
		tpoll.Debugf("omap built with %d elements in %s", len(m.IdxToName), since.String())
	}
	return m, err
}

func (m *OMap) walkCB(pdu gosnmp.SnmpPDU) error {
	idx := pdu.Name[len(m.Oid.Numeric)+2:]
	var ifN string
	ifN, ok := pdu.Value.(string)
	if !ok {
		ifN = string(pdu.Value.([]byte))
	}
	m.IdxToName[idx] = ifN
	m.NameToIdx[ifN] = idx
	return nil
}
