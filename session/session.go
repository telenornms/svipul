package session

import (
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/tpoll"
)

type Session struct {
	S         *gosnmp.GoSNMP
	Target    string
	Community string
}

func (s *Session) init() error {
	gs := gosnmp.GoSNMP{
		Port:               161,
		Transport:          "udp",
		Community:          s.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(3) * time.Second,
		Retries:            3,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
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

func (s *Session) BulkWalk(node tpoll.Node, cb func(pdu gosnmp.SnmpPDU) error) error {
	return s.S.BulkWalk(node.Numeric, cb)
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
