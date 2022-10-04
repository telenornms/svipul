// Copyright 2012 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// This program demonstrates BulkWalk.
package main

import (
	"fmt"
	"os"
	"time"
	"strconv"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/tpoll/smierte"
)

type Session struct {
	S	*gosnmp.GoSNMP
	Target	string
	Community	string
}

func (s *Session) init() error {
	gs := gosnmp.GoSNMP{
		Port: 161,
		Transport: "udp",
		Community: s.Community,
		Version: gosnmp.Version2c,
		Timeout: time.Duration(10) * time.Second,
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
	fmt.Printf("%s = ", pdu.Name)

	switch pdu.Type {
	case gosnmp.OctetString:
		b := pdu.Value.([]byte)
		fmt.Printf("STRING: %s\n", string(b))
	default:
		fmt.Printf("TYPE %d: %d\n", pdu.Type, gosnmp.ToBigInt(pdu.Value))
	}
	return nil
}

type IFMap struct {
	IdxToName map[int]string
	NameToIdx map[string]int
}

func BuildIFMap(s *Session) (*IFMap, error) {
	var ifm IFMap
	ifm.IdxToName = make(map[int]string)
	ifm.NameToIdx = make(map[string]int)
	err := ifm.Populate(s)
	return &ifm, err
}
const ifName = ".1.3.6.1.2.1.31.1.1.1.1"
func (ifm *IFMap) Populate(s *Session) error {
	return s.S.BulkWalk(ifName, ifm.walkCB)
}

func (ifm *IFMap) walkCB(pdu gosnmp.SnmpPDU) error {
	fmt.Printf("index %s is %s\n", pdu.Name[len(ifName)+1:], string(pdu.Value.([]byte)))
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
	mib := smierte.Config{}
	mib.Modules = []string{
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"}
	mib.Paths = []string{"/usr/share/snmp/mibs"}
	mib.Init()
	s, err := NewSession("192.168.122.41")
	
	if err != nil {
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}
	defer s.Finalize()
	ifm, err := BuildIFMap(s)
	if err != nil {
		fmt.Printf("Build Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ifmap: %#v\n", ifm)

	m,err := mib.Lookup(os.Args[1])
	if err != nil {
		fmt.Printf("Lookup Error: %v\n", err)
		os.Exit(1)
	}
	err = s.BulkWalk(m.Numeric)
	if err != nil {
		fmt.Printf("n2: %v\n", m.Numeric)
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}
}

