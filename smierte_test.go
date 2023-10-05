/*
 * svipul SMIerte tests
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

package svipul_test

import (
	"github.com/telenornms/svipul"
	"github.com/telenornms/svipul/smierte"
	"testing"
)

func TestLoading(t *testing.T) {
	err := smierte.Init(svipul.Config.MibModules, svipul.Config.MibPaths)
	if err != nil {
		t.Errorf("failed to load smi modules: %v", err)
	}

	node, err := smierte.Lookup("sysName")
	if err != nil {
		t.Errorf("failed to lookup sysName: %v", err)
	}
	if node.Type == nil {
		t.Errorf("sysName types is nil")
	}
	if node.Numeric != "1.3.6.1.2.1.1.5" {
		t.Errorf("expected node numeric to be `1.3.6.1.2.1.1.5', got: %s", node.Numeric)
	}

	sysName := "1.3.6.1.2.1.1.5"
	sysName543 := "1.3.6.1.2.1.1.5.543"
	sysName123 := "1.3.6.1.2.1.1.5.123"
	node, err = smierte.Lookup("sysName.123")
	if err != nil {
		t.Errorf("failed to lookup sysName.123: %v", err)
	}
	if node.Type == nil {
		t.Errorf("sysName.123 types is nil")
	}
	if node.Numeric != sysName {
		t.Errorf("expected node numeric to be `%s', got: %s", sysName, node.Numeric)
	}
	if node.Qualified != sysName123 {
		t.Errorf("expected node numeric to be `%s', got: %s", sysName, node.Numeric)
	}

	node, err = smierte.Lookup(sysName543)
	if err != nil {
		t.Errorf("failed to lookup %s: %v", sysName543, err)
	}
	if node.Type == nil {
		t.Errorf("%s types is nil", sysName543)
	}
	if node.Numeric != sysName {
		t.Errorf("expected node numeric to be `%s', got: %s", sysName, node.Numeric)
	}
	if node.Qualified != sysName543 {
		t.Errorf("expected node qualified to be `%s', got: %s", sysName543, node.Qualified)
	}
}
