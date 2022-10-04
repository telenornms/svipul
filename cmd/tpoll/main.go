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
	"fmt"
	"os"
	"time"
	"strconv"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/tpoll/smierte"
	"github.com/telenornms/tpoll"
)

type Session struct {
	S	*gosnmp.GoSNMP
	Target	string
	Community	string
	Mib	*smierte.Config
	Ifm	IFMap
}

func (s *Session) init() error {
	gs := gosnmp.GoSNMP{
		Port: 161,
		Transport: "udp",
		Community: s.Community,
		Version: gosnmp.Version2c,
		Timeout: time.Duration(3) * time.Second,
		Retries: 3,
		ExponentialTimeout: true,
		MaxOids: gosnmp.MaxOids,
	}
	gs.Target = s.Target
	err := gs.Connect()
	if err != nil {
		return fmt.Errorf("snmp connect: %w\n", err)
	}
	s.S = &gs
	return nil
}

func (s *Session) Finalize() {
	s.S.Conn.Close()
}

func NewSession(target string) (*Session, error) {
	var s Session
	s.Target = target
	s.Community = "public"
	err := s.init()
	if err != nil {
		return nil, err
	}
	return &s, nil
}


func (s *Session) BulkWalk(oid string) error {
	cb := func(pdu gosnmp.SnmpPDU) error {
		return s.printValue(pdu)
	}
	err := s.S.BulkWalk(oid, cb)
	return err
}

func (s *Session) printValue(pdu gosnmp.SnmpPDU) error {
	var name = pdu.Name
	if s.Mib != nil {
		n,err := s.Mib.Lookup(pdu.Name)
		if err != nil {
			tpoll.Logf("lookup failed: %s", err)
		} else {
			trailer := pdu.Name[len(n.Numeric)+1:]
			if len(trailer) > 1 {
				idxN64,_ := strconv.ParseInt(trailer[1:],10,32)
				idx := int(idxN64)
				
				if s.Ifm.IdxToName[idx] != "" {
					trailer = fmt.Sprintf(".%s", s.Ifm.IdxToName[idx])
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
		tpoll.Logf("%s = TYPE %d: %d\n",name, pdu.Type, gosnmp.ToBigInt(pdu.Value))
	}
	return nil
}

type IFMap struct {
	IdxToName map[int]string
	NameToIdx map[string]int
}

func BuildIFMap(s *Session) (error) {
	var ifm IFMap
	ifm.IdxToName = make(map[int]string)
	ifm.NameToIdx = make(map[string]int)
	err := ifm.Populate(s)
	s.Ifm = ifm
	return err
}

const ifName = ".1.3.6.1.2.1.31.1.1.1.1"
func (ifm *IFMap) Populate(s *Session) error {
	return s.S.BulkWalk(ifName, ifm.walkCB)
}

func (ifm *IFMap) walkCB(pdu gosnmp.SnmpPDU) error {
	idx := pdu.Name[len(ifName)+1:]
	ifN := string(pdu.Value.([]byte))
	idxN64,err := strconv.ParseInt(idx,10,32)
	if err != nil {
		return err
	}
	idxN := int(idxN64)
	ifm.IdxToName[idxN] = ifN
	ifm.NameToIdx[ifN] = idxN
	return nil
}
	
func main() {
	mib := &smierte.Config{}
	mib.Modules = []string{
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"}
	mib.Paths = []string{"/usr/share/snmp/mibs"}
	err := mib.Init()
	if err != nil {
		tpoll.Fatalf("failed to load mibs: %s", err)
	}
	s, err := NewSession("192.168.122.41")
	if err != nil {
		tpoll.Fatalf("failed to start session: %s", err)
	}
	s.Mib = mib
	defer s.Finalize()
	err = BuildIFMap(s)
	if err != nil {
		tpoll.Fatalf("failed to build IF-map: %s", err)
	}

	m,err := mib.Lookup(os.Args[1])
	if err != nil {
		tpoll.Fatalf("unable to lookup mib/oid/thingy: %s", err)
	}
	err = s.BulkWalk(m.Numeric)
	if err != nil {
		tpoll.Fatalf("Walk Failed: %v\n", err)
	}
}

