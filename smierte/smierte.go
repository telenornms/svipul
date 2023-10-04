/*
 * tpoll smi-pain
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

package smierte

/*
Package smierte handles loading MIB files and modules (SMI)-stuff. The name
is a play on SMI and smerte (pain), because this is such a painful process.

While this is based on gosmi, we should try to hide as much as that as
possible because it's not unlikely that it'll be switched.
*/

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
	"github.com/telenornms/svipul"
)

var lock sync.Mutex

// cache is an internal OID-cache for Nodes, to avoid expensive SMI-lookups
// for what is most likely very repetitive lookups. So far, extremely
// simple with no LRU or anything.
var cache sync.Map

// Init loads MIB files from disk and a hard-coded list of modules
func Init(inmodules []string, paths []string) error {
	gosmi.Init()

	for _, path := range paths {
		tpoll.Debugf("mib path added: %s", path)
		gosmi.AppendPath(path)
	}
	loaded := 0
	var modules []string
	for _, module := range inmodules {
		moduleName, err := gosmi.LoadModule(module)
		if err != nil {
			return fmt.Errorf("module load for %s failed: %w", moduleName, err)
		}
		modules = append(modules, moduleName)
		loaded++
	}
	tpoll.Debugf("Loaded SMI modules: %s", strings.Join(modules, ", "))
	tpoll.Logf("Loaded %d SMI modules", loaded)
	return nil
}

// Lookup looks up an oid, first in cache, then regularly. It needs to do a
// bit of lock-juggling, despite cache being sync.Map: the sync.Map bit is
// safe enough, but there's a good chance we'll do multiple lookups in
// parallel here. Could probably be simplified, need to benchmark how slow
// types.OidFromString, GetNodeByOID and GetNode is...
func Lookup(item string) (tpoll.Node, error) {
	if chit, ok := cache.Load(item); ok {
		cast, _ := chit.(*tpoll.Node)
		return *cast, nil
	}
	lock.Lock()
	defer lock.Unlock()
	// Re-check in case other go-routine got it
	if chit, ok := cache.Load(item); ok {
		cast, _ := chit.(*tpoll.Node)
		return *cast, nil
	}
	var ret tpoll.Node
	defer cache.Store(item, &ret)
	ret.Key = item
	match, _ := regexp.Match("^[0-9.]+$", []byte(item))
	var err error
	var n gosmi.SmiNode
	endy := ""
	if match {
		oid, err := types.OidFromString(item)
		if err != nil {
			return ret, fmt.Errorf("unable to resolve OID to string: %w", err)
		}
		ret.Lookedup = false
		n, err = gosmi.GetNodeByOID(oid)
	} else {
		ret.Lookedup = true
		// This is to allow looking up sysName.0 - it's probably
		// not very close to perfect.
		splitty := strings.Split(item, ".")
		n, err = gosmi.GetNode(splitty[0])
		if len(splitty) > 1 {
			endy = "." + strings.Join(splitty[1:], ".")
		}
	}
	if err != nil {
		return ret, fmt.Errorf("gosmi.GetNode failed: %w", err)
	}
	ret.Numeric = n.RenderNumeric() + endy
	ret.Name = n.Render(types.RenderName)
	if n.Type != nil {
		ret.Format = n.Type.Format
	}
	ret.Type = n.Type
	if match {
		ret.Qualified = item[1:]
	}
	return ret, nil
}
