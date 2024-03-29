/*
 * svipul common data definitions
 *
 * Copyright (c) 2023 Telenor Norge AS
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

package svipul

import (
	"github.com/gosnmp/gosnmp"
	"github.com/sleepinggenius2/gosmi/models"
)

// Node is a rendered SMI node, e.g.: the result of a lookup. Usually
// handled by the smierte sub-package, but needs to be defined up here to
// avoid circular dependencies
type Node struct {
	Key       string // original input key, kept for posterity
	Name      string // sysName
	Numeric   string // .1.3.6.1.2.1.1.1 // root
	Qualified string // .1.3.6.1.2.1.1.1.0 // "full", with index
	Format    string // 255a <-- possibly obsolete
	Type      *models.Type
	Lookedup  bool // True if key was not a pure OID
}

// Walker is an interface for performing a BulkWalk, without having to
// worry about the underlying session. Today, only the session-subpackage
// and session.Session type implements it. Since it's tied to both a Node
// and a gosnmp.SnmpPDU type, it's rather strongly connected to SNMP atm.
type Walker interface {
	BulkWalk(node []Node, cb func(pdu gosnmp.SnmpPDU) error) error
}
