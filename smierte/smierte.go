package smierte

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)


type Config struct {
	Modules []string // SMI modules to load
	Paths []string // Paths to the modules
}


// Node is the rendered 
type Node struct {
	Key	string	// original input key, kept for posterity
	Name	string
	Numeric	string // I KNOW
	Qualified string
}

// cache is an internal OID-cache for Nodes, to avoid expensive SMI-lookups
// for what is most likely very repetitive lookups. So far, extremely
// simple with no LRU or anything.
var cache sync.Map

// Init loads MIB files from disk and a hard-coded list of modules
func (c *Config) Init() {
	gosmi.Init()

	modules := []string {
		"SNMPv2-MIB",
		"ENTITY-MIB",
		"IF-MIB",
		"IP-MIB",
		"IP-FORWARD-MIB"}

	for i, module := range modules {
		moduleName, err := gosmi.LoadModule(module)
		if err != nil {
			fmt.Printf("Init Error: %s\n", err)
			return
		}
		fmt.Printf("Loaded module %s\n", moduleName)
		modules[i] = moduleName
	}
}


func (c *Config) Lookup(item string) (Node, error) {
	if chit, ok := cache.Load(item); ok {
		cast,_ := chit.(*Node)
		fmt.Printf("Cache hit\n")
		return *cast, nil
	}
	var ret Node
	// We set this early because there's currently no reason to assume
	// a cache miss will magically become a cache hit later.
	// XXX: When we DO deal with internal reloading, we need to nuke
	// this cache.
	cache.Store(item, &ret)
	ret.Key = item
	match,_ := regexp.Match("^[0-9.]+$", []byte(item))
	var err error
	var n gosmi.SmiNode
	if match {
		oid,err := types.OidFromString(item)
		if err != nil {
			return ret, fmt.Errorf("unable to resolve OID to string: %w", err)
		}
		n, err = gosmi.GetNodeByOID(oid)
	} else {
		n, err = gosmi.GetNode(item)
	}
	if err != nil {
		return ret, fmt.Errorf("gosmi.GetNode failed: %w", err)
	}
	ret.Numeric = n.RenderNumeric()
	ret.Name = n.Render(types.RenderName)
	ret.Qualified = n.RenderQualified()
	for i := types.Render(1); i<7; i++ {
		fmt.Printf("render %d: %s\n", i, n.Render(i))
	}
	fmt.Printf("noden: %#v\n", n.Node)
	return ret, nil
}
