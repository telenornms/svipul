/*
 * tpoll inventory
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

/*
Package inventory deals with inventory locking and syncing.

Today, this is mostly a dummy-package, but the intention is to sync it with
a central database.
*/
package inventory

import (
	"fmt"
	"github.com/telenornms/svipul"
	"sync"
)

var targets sync.Map

type Host struct {
	Address   string
	Community string
}

// LockHost acquires a host-level lock and relevant credentials. Must call
// h.Unlock() when done.
func LockHost(t string) (Host, error) {
	h := Host{}
	_, loaded := targets.LoadOrStore(t, 1)
	if loaded {
		return h, fmt.Errorf("target still locked, refusing to start more runs")
	}
	h.Address = t
	h.Community = tpoll.Config.DefaultCommunity
	return h, nil
}

// Unlock releases the host-level lock.
func (h *Host) Unlock() {
	targets.Delete(h.Address)
}
