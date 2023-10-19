/*
 * svipul test program
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
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/telenornms/skogul"
	sconfig "github.com/telenornms/skogul/config"
	"github.com/telenornms/svipul"
	//	"github.com/sleepinggenius2/gosmi/models"
	"github.com/sleepinggenius2/gosmi/types"
	"github.com/telenornms/svipul/inventory"
	"github.com/telenornms/svipul/omap"
	"github.com/telenornms/svipul/session"
	"github.com/telenornms/svipul/smierte"
)

// Task is tied to a single SNMP run/walk and a single host
type Task struct {
	OMap   *omap.OMap    // Engine populates uniquely for each target
	Metric skogul.Metric // New metric for each run.
	Result ResolveM
}

// Engine is semi-global state for SNMP, including a "cached" OMap ... map
type Engine struct {
	Skogul *sconfig.Config                  // output
	OMap   map[string]map[string]*omap.OMap // Caches/stores looked up/built omaps
}

// Init reads configuration and whatnot for the engine
func (e *Engine) Init(sc string) error {
	var err error
	e.Skogul, err = sconfig.Path(sc)
	if err != nil {
		return fmt.Errorf("skogul-config failed loading: %w", err)
	}
	if e.Skogul.Handlers["svipul"] == nil {
		return fmt.Errorf("missing svipul handler in skogul config")
	}
	e.OMap = make(map[string]map[string]*omap.OMap)
	err = smierte.Init(svipul.Config.MibModules, svipul.Config.MibPaths)
	if err != nil {
		svipul.Fatalf("failed to load mibs: %s", err)
	}
	return nil
}

// GetOmap builds an omap on demand, or returns an already built one
func (e *Engine) GetOmap(target string, key string, sess *session.Session) (*omap.OMap, error) {
	var err error
	if e.OMap[target][key] != nil {
		if time.Since(e.OMap[target][key].Timestamp) > svipul.Config.MaxMapAge {
			svipul.Logf("Deleting aged out omap for %s", target)
			e.OMap[target][key] = nil
		} else {
			return e.OMap[target][key], nil
		}
	}
	o, err := omap.BuildOMap(sess, key)
	if err != nil {
		return nil, fmt.Errorf("failed to build IF-map: %w", err)
	}
	if e.OMap[target] == nil {
		e.OMap[target] = make(map[string]*omap.OMap)
	}
	e.OMap[target][key] = o
	return e.OMap[target][key], nil
}

// ClearOmap clears/nukes/empties the map cache for a target/key combo. If
// the key is blank, ALL maps for that target is cleared.
func (e *Engine) ClearOmap(target string, key string) error {
	if key == "" {
		svipul.Logf("Deleting all maps for %s on request", target)
		delete(e.OMap, target)
		return nil
	}
	if e.OMap[target] != nil {
		svipul.Logf("Deleting `%s'-map for %s on request", key, target)
		delete(e.OMap[target], key)
		return nil
	}
	svipul.Logf("Map `%s' for %s not found while trying to clear chache. Nothing to do. Wohoo!", key, target)
	return nil
}

// Run starts an SNMP session for a target and collects the specified oids,
// if emap is true, it will use an oid/element map, building it on demand.
//
// TODO: This needs to be split up and possibly refactored. It's a bit of a
// beast.
func (e *Engine) Run(o Order) error {
	host, err := inventory.LockHost(o.Target)
	if err != nil {
		return fmt.Errorf("unable to acquire host lock: %w", err)
	}
	defer host.Unlock()
	if o.Mode == ClearMap {
		return e.ClearOmap(o.Target, o.Key)
	}
	if o.Elements != nil && len(o.Elements) > 0 {
		if o.Key == "" {
			svipul.Debugf("elements provided, but not key. Assuming ifName")
			o.Key = "ifName"
		}
	}

	community := host.Community
	if o.Community != "" {
		community = o.Community
	}
	sess, err := session.NewSession(o.Target, community)
	if err != nil {
		return fmt.Errorf("session creation failed: %w", err)
	}
	defer sess.Finalize()
	svipul.Debugf("%s - starting run", o.Target)

	if o.Mode == BuildMap {
		if o.Key == "" {
			svipul.Debugf("Requested building of a map, but no key provided. Assuming ifName")
			o.Key = "ifName"
		}

		err := e.ClearOmap(o.Target, o.Key)
		if err != nil {
			return fmt.Errorf("unable to clear omap: %w", err)
		}
		_, err = e.GetOmap(o.Target, o.Key, sess)
		if err != nil {
			return fmt.Errorf("unable to build omap: %w", err)
		}
		return nil
	}

	t := Task{}
	if o.Key != "" {
		t.OMap, err = e.GetOmap(o.Target, o.Key, sess)
		if err != nil {
			return fmt.Errorf("failed to build IF-map: %w", err)
		}
	}

	lookedup := false
	m := make([]svipul.Node, 0, len(o.Oids))
	for _, arg := range o.Oids {
		nym, err := smierte.Lookup(arg)
		if err != nil {
			return fmt.Errorf("unable to look up oid: %w", err)
		}
		m = append(m, nym)
		if nym.Lookedup {
			lookedup = true
		}
	}
	if o.Result == Auto {
		if lookedup {
			t.Result = Resolve
		} else {
			t.Result = OID
		}
	} else {
		t.Result = o.Result
	}
	if len(m) < 1 {
		return fmt.Errorf("trying to start rul with 0 oids?")
	}
	t.Metric.Metadata = make(map[string]interface{})
	t.Metric.Metadata["target"] = o.Target
	if o.ID != "" {
		t.Metric.Metadata["id"] = o.ID
	}
	t.Metric.Data = make(map[string]interface{})
	if o.Mode == GetElements {
		nym := make([]svipul.Node, 0, len(m)*len(o.Elements))
		for _, oid := range m {
			for _, e := range o.Elements {
				for idx, einner := range t.OMap.NameToIdx {
					match, _ := regexp.Match(e, []byte(idx))
					if match {
						eid := einner
						nynode := oid
						nynode.Qualified = nynode.Qualified + fmt.Sprintf(".%s", eid)
						nym = append(nym, nynode)
					}

				}
			}
		}
		err = sess.Get(nym, t.bwCB)
	} else if o.Mode == Walk {
		err = sess.BulkWalk(m, t.bwCB)
	} else if o.Mode == Get {
		err = sess.Get(m, t.bwCB)
	} else {
		return fmt.Errorf("unsupported mode")
	}
	if err != nil {
		return fmt.Errorf("snmp get/walk failed: %w", err)
	}
	c := skogul.Container{}
	c.Metrics = append(c.Metrics, &t.Metric)

	err = e.Skogul.Handlers["svipul"].Handler.TransformAndSend(&c)
	if err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	return nil
}

// saveNode stores a result
func (t *Task) saveNode(pdu gosnmp.SnmpPDU, v interface{}) error {
	if t.Result == OID {
		t.Metric.Data[pdu.Name] = v
		return nil
	}
	var name = pdu.Name
	var element = ""

	n, err := smierte.Lookup(pdu.Name)
	if err != nil {
		svipul.Logf("lookup failed: %s", err)
	} else {
		var trailer string
		if len(n.Numeric) >= len(pdu.Name)-1 || (n.Qualified != "" && pdu.Name == n.Qualified[1:]) {
			svipul.Logf("trailer-issues: %s vs %v", n.Numeric, pdu)
			trailer = "0"
		} else {
			trailer = pdu.Name[len(n.Numeric)+1:][1:]
			if len(trailer) > 0 {
				if t.OMap != nil && t.OMap.IdxToName[trailer] != "" {
					trailer = t.OMap.IdxToName[trailer]
				}
			}
		}
		name = n.Name
		element = trailer
	}

	if t.Metric.Data[element] == nil {
		t.Metric.Data[element] = make(map[string]interface{})
	}
	(t.Metric.Data[element].(map[string]interface{}))[name] = v
	return nil
}

// bwCB is the callback used for each PDU received during an SNMP GET of
// some sort. It just "decodes" the value and triggers storage.
//
// The decoding is a bit finnicky. We only want to use the "formatted"
// stuff for types that require rendering, while numbers should be left
// intact. And then there's OctetString where we DO want to use DisplayHint
// if present, but NOT if it isn't present, because the default is
// atrocious.
func (t *Task) bwCB(pdu gosnmp.SnmpPDU) error {
	var v interface{}
	node, err := smierte.Lookup(pdu.Name)
	if err != nil {
		svipul.Logf("PDU/Node lookup failed during callback: %v", err)
	}
	if node.Type == nil {
		return t.saveNode(pdu, pdu.Value)
	}
	foo := node.Type.FormatValue(pdu.Value)
	if node.Type.BaseType == types.BaseTypeUnknown ||
		node.Type.BaseType == types.BaseTypeObjectIdentifier ||
		node.Type.BaseType == types.BaseTypeEnum ||
		node.Type.BaseType == types.BaseTypeBits ||
		node.Type.BaseType == types.BaseTypePointer {
		v = foo.Formatted
	} else if node.Type.BaseType == types.BaseTypeOctetString {
		if node.Type.Format == "" {
			switch foo.Raw.(type) {
			case string:
				v = foo.Raw
			case []uint8:
				// This one is a bit iffy, since I
				// don't know if there are
				// problematic octet strings out
				// there, but I _do_ know
				// hrSWInstalledName will fail to
				// render sensibly without it.
				v = string(foo.Raw.([]uint8))
			default:
				v = foo.Formatted
			}
		} else {
			v = foo.Formatted
		}
	} else {
		v = pdu.Value
	}

	return t.saveNode(pdu, v)
}

// Order is the central object for kicking Svipul into action. An order
// always operates on a target (a host/switch, either IP address or
// hostname) and using a mode. Depending on the mode, Svipul can either
// request OIDS from the target system, build table/element maps or clear
// the map cache. There are more than one method of getting OIDS.
//
// OIDs can be provided either as a list of numeric IDs, or by the symbolic
// names. E.g.: .1.3.6.1.2.1.1.5.0 is valid, but so is ifHCInOctets. At the
// time of this writing, ifHCInOctets.10 is NOT valid, but that is planned
// for the future.
//
// If the Elements array is populated, an element map will be used to fetch
// oids for the matching elements. More plainly: Elements can match
// interface names and then Svipul will build up GET requests for the
// provided OIDs for each index.
//
// If Key is provided, that is used as the Key to build an element map. By
// default, ifName is used, making the defaults suitable for looking up
// OIDS under ifTable and ifXTable.
//
// Community is the community to use to connect to the host.
//
// ID is an optional identification which is not used by Svipul at all, but
// included in the result to allow a caller to match the order to the
// result.
//
// Result determines how the result is formatted. By default, it will try
// to match the input. E.g.: If numeric OIDs were used in the input, that's
// used in the output. If symbolic names were used, that's used for the
// result by default. This behavior can be overridden by providing "oid" to
// leave OIDs unresolved and "Resolve" to attempt to always resolve them.
type Order struct {
	Target    string   // Host/target
	Oids      []string // OIDs, also accepts logical names (e.g.: ifName)
	Elements  []string // Elemnts, if GetElements mode. Elements == interfaces (could be other in the future)
	Key       string   // Map key to use for looking up elements
	Mode      Mode     // What mode to use
	Community string   `json:",omitempty"` // Community to use, blank == figure it out yourself/use default (meaning depends on issuer)
	ID        string   `json:",omitempty"`
	Result    ResolveM // Auto (default) = resolve based on input, OID = leave OIDs unresolved, Resolve = try to resolve
	delivery  amqp.Delivery
}

func (o Order) String() string {
	return o.Target
}

type ResolveM int

const (
	Auto ResolveM = iota
	OID
	Resolve
)

func (r *ResolveM) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(s)
	switch s {
	case "auto":
		*r = Auto
	case "oid":
		*r = OID
	case "resolve":
		*r = Resolve
	default:
		return fmt.Errorf("invalid resolver mode: %s", s)
	}
	return nil
}

func (r ResolveM) MarshalJSON() ([]byte, error) {
	switch r {
	case Auto:
		return []byte("\"Auto\""), nil
	case OID:
		return []byte("\"OID\""), nil
	case Resolve:
		return []byte("\"Resolve\""), nil
	default:
		return []byte("\"\""), fmt.Errorf("invalid resolve mode %d!", r)
	}
}

type Mode int

const (
	Walk        Mode = iota // Do a walk
	Get                     // Get just these oids
	GetElements             // Get these specific oids, but per elements
	BuildMap                // Build an OMap
	ClearMap                // Clear the OMap cache
)

func (m *Mode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(s)
	switch s {
	case "walk":
		*m = Walk
	case "get":
		*m = Get
	case "getelements":
		*m = GetElements
	case "buildmap":
		*m = BuildMap
	case "clearmap":
		*m = ClearMap
	default:
		return fmt.Errorf("invalid mode: %s", s)
	}
	return nil
}

func (m Mode) MarshalJSON() ([]byte, error) {
	switch m {
	case Walk:
		return []byte("\"Walk\""), nil
	case Get:
		return []byte("\"Get\""), nil
	case GetElements:
		return []byte("\"GetElements\""), nil
	case BuildMap:
		return []byte("\"BuildMap\""), nil
	case ClearMap:
		return []byte("\"ClearMap\""), nil
	default:
		return []byte("\"\""), fmt.Errorf("invalid mode %d!", m)
	}
}

func (e *Engine) Listener(c chan Order, name string) {
	svipul.Debugf("Starting listener %s...", name)
	for order := range c {
		now := time.Now()
		err := e.Run(order)
		since := time.Since(now).Round(time.Millisecond * 10)
		if err != nil {
			requeue := true
			if order.delivery.Redelivered {
				requeue = false
			}
			svipul.Logf("[%2s]: %-15s FAIL %s: %s (requeue: %v)", name, order, since.String(), err, requeue)
			if requeue {
				delayR := rand.Int() % 10
				d := time.Second*1 + time.Second*time.Duration(delayR)
				svipul.Debugf("Sleeping %v before NACK/requeue", d)
				time.Sleep(d)
			}
			err2 := order.delivery.Nack(false, requeue)
			if err2 != nil {
				svipul.Logf("NAck failed: %s", err2)
			}

		} else {
			svipul.Logf("[%2s]: %-15s OK %s", name, order, since.String())
			err2 := order.delivery.Ack(false)
			if err2 != nil {
				svipul.Logf("Ack failed: %s", err2)
			}
		}
	}
}

func main() {
	var configFile string
	flag.BoolVar(&svipul.Config.Debug, "debug", false, "enable debug")
	flag.StringVar(&configFile, "f", "/etc/svipul/snmp.toml", "snmp config file")
	flag.Parse()
	if err := svipul.ParseConfig(configFile); err != nil {
		svipul.Fatalf("Couldn't parse config: %s", err)
	}
	svipul.Debugf("Read config file: %s", configFile)
	svipul.Init()
	e := Engine{}
	err := e.Init(svipul.Config.OutputConfig)
	if err != nil {
		svipul.Fatalf("Couldn't initialize engine: %s", err)
	}
	c := make(chan Order, 0)
	for i := 0; i < svipul.Config.Workers; i++ {
		go e.Listener(c, fmt.Sprintf("%d", i))
		time.Sleep(time.Microsecond * 20)
	}
	svipul.Logf("Started %d workers", svipul.Config.Workers)
	amUrl, err := url.Parse(svipul.Config.Broker)
	if err != nil {
		svipul.Fatalf("Can't parse broker url: %s", err)
	}
	svipul.Debugf("Connecting to broker: %v", amUrl.Redacted())
	conn, err := amqp.Dial(svipul.Config.Broker)
	if err != nil {
		svipul.Fatalf("can't connect to broker: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		svipul.Fatalf("can't get channel: %s", err)
	}
	defer ch.Close()
	err = ch.Qos(svipul.Config.Workers+1, 0, true)
	if err != nil {
		svipul.Fatalf("can't set qos: %s", err)
	}

	q, err := ch.QueueDeclare(
		"svipul", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		svipul.Fatalf("can't declare queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		svipul.Fatalf("can't register consumer: %s", err)
	}
	svipul.Logf("Listening for orders")
	for d := range msgs {
		order := Order{}
		err = json.Unmarshal(d.Body, &order)
		if err != nil {
			svipul.Logf("order json unmarshal: %s", err)
			d.Reject(false)
			continue
		}
		order.delivery = d
		c <- order
	}
	svipul.Logf("Reached the end. Connection probably dead. Some day, we'll handle this, but not today.")
}
