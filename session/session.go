package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/svipul"
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
		Retries:            1,
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

// Get uses SNMP Get to fetch precise OIDs. it will split it into
// multiple requests if there are more nodes than max-oids.a
func (s *Session) Get(nodes []tpoll.Node, cb func(pdu gosnmp.SnmpPDU, node tpoll.Node) error) error {
	if len(nodes) < 1 {
		return fmt.Errorf("refusing to carry out GET for 0 nodes")
	}
	oids := make([]string, 0, len(nodes))
	originals := make([]string, 0, len(nodes))
	mapback := make(map[string]tpoll.Node)
	for _, a := range nodes {
		on := a.Numeric
		if a.Qualified != "" {
			on = a.Qualified
		}
		numeric := fmt.Sprintf(".%s", on)
		oids = append(oids, numeric)
		originals = append(originals, numeric)
		mapback[numeric] = a
	}
	if len(oids) < 1 || oids[0] == "." || originals[0] == "." {
		return fmt.Errorf("corrupt oid-lookup, probably a bug. oids[0] is blank: nodes: %#v", nodes)
	}
	runs := 0
	for i := 0; i < len(oids); i += 50 {
		end := i + 50
		if end > len(oids) {
			end = len(oids)
		}
		err := s.get(oids[i:end], mapback, cb)
		if err != nil {
			return fmt.Errorf("oid get failed: %w", err)
		}
		runs++
	}
	tpoll.Debugf("run for %d oids finished in %d iterations", len(oids), runs)
	return nil
}

func (s *Session) get(oids []string, mapback map[string]tpoll.Node, cb func(pdu gosnmp.SnmpPDU, node tpoll.Node) error) error {
	originals := oids
	result, err := s.S.Get(oids)
	if err != nil {
		return fmt.Errorf("Get failed: %w", err)
	}
	if result.Error != gosnmp.NoError {
		return fmt.Errorf("response error: %s", result.Error)
	}
	for _, pdu := range result.Variables {
		if pdu.Type == gosnmp.EndOfMibView {
			return fmt.Errorf("issues with pdu (oids: %v), type: %v, pdu: %v", oids, pdu.Type, pdu)
		} else if pdu.Type == gosnmp.NoSuchObject || pdu.Type == gosnmp.NoSuchInstance {
			tpoll.Debugf("got no such object/no such instance when looking for oid. Ignoring. pdu: %v", pdu)
			continue
		}
		found := false
		for _, o := range originals {
			if strings.HasPrefix(pdu.Name, o) {
				found = true
				break
			}
		}
		if found {
			err = cb(pdu, mapback[pdu.Name])
			if err != nil {
				return fmt.Errorf("callback returned error: %w", err)
			}
		} else {
			tpoll.Logf("Invalid pdu returned? WAT: %s", pdu.Name)
		}

	}
	return nil
}

// BulkWalk uses SNMP GetBulk to fetch one or more column/table, calling cb
// for each pdu received.
func (s *Session) BulkWalk(nodes []tpoll.Node, cb func(pdu gosnmp.SnmpPDU, node tpoll.Node) error) error {
	oids := make([]string, 0, len(nodes))
	originals := make([]string, 0, len(nodes))
	for _, a := range nodes {
		numeric := fmt.Sprintf(".%s", a.Numeric)
		oids = append(oids, numeric)
		originals = append(originals, numeric)
	}
	iterations := 0
	misses := 0
	hits := 0
	if oids[0] == "." || originals[0] == "." {
		return fmt.Errorf("corrupt oid-lookup, probably a bug. oids[0] is blank")
	}
	for ; len(oids) > 0; iterations++ {
		revmap := make(map[string]string)
		result, err := s.S.GetBulk(oids, 0, 10)
		if err != nil {
			return fmt.Errorf("GetBulk failed after %d iterations: %w", iterations, err)
		}
		oids = make([]string, 0, 5)
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
				err = cb(pdu, tpoll.Node{})
				if err != nil {
					return fmt.Errorf("callback returned error: %w", err)
				}
				hits++
			}
		}
		for _, r := range revmap {
			oids = append(oids, r)
		}
	}
	tpoll.Debugf("BulkWalk for %d oids done in %d iterations with %d misses and %d hits", len(nodes), iterations, misses, hits)
	return nil
}

func NewSession(target string, community string) (*Session, error) {
	var s Session
	s.Target = target
	s.Community = community
	err := s.init()
	if err != nil {
		return nil, err
	}
	return &s, nil
}
