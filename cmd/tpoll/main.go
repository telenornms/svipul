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
	"strconv"
	"time"
	"sync"

	"github.com/gosnmp/gosnmp"
	"github.com/telenornms/skogul"
	sconfig "github.com/telenornms/skogul/config"
	"github.com/telenornms/tpoll"
	"github.com/telenornms/tpoll/config"
	"github.com/telenornms/tpoll/omap"
	"github.com/telenornms/tpoll/session"
	"github.com/telenornms/tpoll/smierte"
)

// Task is tied to a single SNMP run/walk and a single host
type Task struct {
	OMap   *omap.OMap // Engine populates uniquely for each target
	Mib    *smierte.Config // Engine populates, but same instance
	Metric skogul.Metric // New metric for each run.
}

// Engine is semi-global state for SNMP, including a "cached" OMap ... map
type Engine struct {
	Skogul *sconfig.Config // output
	Mib    *smierte.Config // MIB
	Targets sync.Map
	OMap   map[string]*omap.OMap // Caches/stores looked up/built omaps
}

// Init reads configuration and whatnot for the engine
func (e *Engine) Init(sc string) error {
	var err error
	e.Skogul, err = sconfig.Path(sc)
	if err != nil {
		return fmt.Errorf("skogul-config failed loading: %w", err)
	}
	if e.Skogul.Handlers["tpoll"] == nil {
		return fmt.Errorf("missing tpoll handler in skogul config")
	}
	e.OMap = make(map[string]*omap.OMap)
	mib := &smierte.Config{}
	mib.Paths = config.MibPaths
	mib.Modules = config.MibModules
	err = mib.Init()
	if err != nil {
		tpoll.Fatalf("failed to load mibs: %s", err)
	}
	e.Mib = mib
	return nil
}

// GetOmap builds an omap on demand, or returns an already built one
func (e *Engine) GetOmap(target string, sess *session.Session) (*omap.OMap, error) {
	var err error
	if e.OMap[target] != nil {
		return e.OMap[target], nil
	}
	e.OMap[target], err = omap.BuildOMap(sess, e.Mib, "ifName")
	if err != nil {
		return nil, fmt.Errorf("failed to build IF-map: %w", err)
	}
	return e.OMap[target], nil
}

// Run starts an SNMP session for a target and collects the specified oids,
// if emap is true, it will use an oid/element map, building it on demand.
func (e *Engine) Run(target string, oids []string, emap bool) error {
	_, loaded := e.Targets.LoadOrStore(target, 1)
	if loaded {
		return fmt.Errorf("target still locked, refusing to start more runs")
	}
	defer e.Targets.Delete(target)
	sess, err := session.NewSession(target)
	if err != nil {
		return fmt.Errorf("session creation failed: %w", err)
	}
	defer sess.Finalize()

	t := Task{}
	if emap {
		t.OMap, err = e.GetOmap(target, sess)
		if err != nil {
			return fmt.Errorf("failed to build IF-map: %w", err)
		}
	}
	t.Mib = e.Mib
	m := make([]tpoll.Node, 0, len(os.Args)-1)
	for _, arg := range oids {
		nym, err := e.Mib.Lookup(arg)
		if err != nil {
			fmt.Errorf("unable to look up oid: %w", err)
		}
		m = append(m, nym)
	}
	if len(m) < 1 {
		return fmt.Errorf("trying to start rul with 0 oids?")
	}
	t.Metric.Metadata = make(map[string]interface{})
	t.Metric.Metadata["oids"] = oids
	t.Metric.Metadata["host"] = target
	t.Metric.Metadata["useMap"] = emap
	t.Metric.Data = make(map[string]interface{})
	err = sess.BulkWalk(m, t.bwCB)
	if err != nil {
		return fmt.Errorf("walk failed: %w", err)
	}
	c := skogul.Container{}
	c.Metrics = append(c.Metrics, &t.Metric)

	err = e.Skogul.Handlers["tpoll"].Handler.TransformAndSend(&c)
	if err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	return nil
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

				if t.OMap != nil && t.OMap.IdxToName[idx] != "" {
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
	case gosnmp.Boolean:
		b := pdu.Value.(bool)
		(t.Metric.Data[element].(map[string]interface{}))[name] = b
	default:
		(t.Metric.Data[element].(map[string]interface{}))[name] = gosnmp.ToBigInt(pdu.Value)
		//(t.Metric.Data[element].(map[string]interface{}))[name+"Type"] = fmt.Sprintf("0x%02X", byte(pdu.Type))
	}
	return nil
}

type Order struct {
	Target string
	Oids	[]string
	EMap	bool
}

func (e *Engine) Listener(c chan Order, name string) {
	tpoll.Logf("Starting listener %s...", name)
	for order := range c {
		now := time.Now()
		err := e.Run(order.Target, order.Oids, order.EMap)
		since := time.Since(now).Round(time.Millisecond * 10)


		if err != nil {
			tpoll.Logf("%s[%s]: %s failed: %s", name, since.String(), order.Target, err)
		} else {
			tpoll.Logf("%s[%s]: %s OK", name, since.String(), order.Target)
		}
	}
}

func main() {
	e := Engine{}
	err := e.Init("skogul")
	if err != nil {
		tpoll.Fatalf("Couldn't initialize engine: %s", err)
	}
	c := make(chan Order, 1)
	for i := 0; i < 10; i++ {
		go e.Listener(c, fmt.Sprintf("%d", i))
	}
	for {
		c <- Order{"192.168.122.41", os.Args[1:], true}
		c <- Order{"192.168.2.3", os.Args[1:], true}
//		c <- Order{"192.168.2.3", os.Args[1:], true}
//		c <- Order{"192.168.2.3", os.Args[1:], true}
//		c <- Order{"192.168.122.41", os.Args[1:], false}
		c <- Order{"192.168.2.3", os.Args[1:], false}
		time.Sleep(time.Second * 5) 
	}
}
