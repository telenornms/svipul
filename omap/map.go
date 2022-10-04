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
	"strconv"
	"github.com/telenornms/tpoll"
	"github.com/telenornms/tpoll/smierte"
	"github.com/gosnmp/gosnmp"
)

// OMap is a two-way map of index to name, the typical case is ifIndex to
// ifName, but can be anything.
type OMap struct {
	IdxToName map[int]string
	NameToIdx map[string]int
	Oid	  tpoll.Node	// OID used to build the map, e.g.: ifName
}

func BuildOMap(w tpoll.Walker, mib *smierte.Config, oid string) (*OMap,error) {
	m := &OMap{}
	var err error
	m.IdxToName = make(map[int]string)
	m.NameToIdx = make(map[string]int)
	m.Oid, err = mib.Lookup(oid)
	if err != nil {
		return nil, fmt.Errorf("lookup of oid %s failed: %w", oid, err)
	}
	err = w.BulkWalk(m.Oid, m.walkCB)
	return m,err
}

func (m *OMap) walkCB(pdu gosnmp.SnmpPDU) error {
	idx := pdu.Name[len(m.Oid.Numeric)+2:]
	ifN := string(pdu.Value.([]byte))
	idxN64, err := strconv.ParseInt(idx, 10, 32)
	if err != nil {
		return err
	}
	idxN := int(idxN64)
	m.IdxToName[idxN] = ifN
	m.NameToIdx[ifN] = idxN
	return nil
}
