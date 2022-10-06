package session

import (
	"fmt"
	"time"
	"strings"

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

func (s *Session) BulkWalk(nodes []tpoll.Node, cb func(pdu gosnmp.SnmpPDU) error) error {
	oids := make([]string,0, len(nodes))
	originals := make([]string,0, len(nodes))
	for _,a := range nodes {
		numeric := fmt.Sprintf(".%s", a.Numeric)
		oids = append(oids, numeric)
		originals = append(originals, numeric)
	}
	iterations := 0
	misses := 0
	for ; len(oids) > 0; iterations++{
		revmap := make(map[string]string)
		result, err := s.S.GetBulk(oids, 0, 3)
		oids = make([]string, 0, 5)
		if err != nil {
			return err
		}
		if result.Error != gosnmp.NoError {
			return fmt.Errorf("response error: %s", result.Error)
		}
		for _, pdu := range result.Variables {
			if pdu.Type == gosnmp.EndOfMibView || pdu.Type == gosnmp.NoSuchObject || pdu.Type == gosnmp.NoSuchInstance {
				return fmt.Errorf("walk issues with pdu, type: %v", pdu.Type)
			}
			found := false
			for _, o := range originals {
				if strings.HasPrefix(pdu.Name, o+".") {
					found = true
					revmap[o] = pdu.Name
					break
				}
			}
			if !found {
				misses++
			} else {
				err = cb(pdu)
				if err != nil {
					return fmt.Errorf("callback returned error: %w", err)
				}
			}
		}
		for _,r := range revmap {
			oids = append(oids, r)
		}
	}
	tpoll.Debugf("BulkWalk for %d oids done in %d iterations with %d misses", len(nodes), iterations, misses)
	return nil
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
