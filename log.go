/*
 * tpoll log-wrappers
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

package tpoll

/*
log.go is largely a wrapper around log for now, mainly so I can start doing
regular calls to log without having to worry about future-proofing it.

Add wrappers on demand.

The one concession it has is that it adds Debug/Debugf which evaluates if
we've turned on debugging. This makes calls to tpoll.Debug() very fast
when it's disabled.
*/

import (
	"github.com/telenornms/tpoll/config"
	"log"
)

func init() {
	d := log.Default()
	d.SetFlags(log.Ltime)
}

func Log(v ...any) {
	log.Print(v...)
}

func Logf(format string, v ...any) {
	log.Printf(format, v...)
}

func Logln(v ...any) {
	log.Println(v...)
}

func Fatal(v ...any) {
	log.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

func Fatalln(v ...any) {
	log.Fatalln(v...)
}

func Debug(v ...any) {
	if config.Debug {
		Log(v...)
	}
}

func Debugf(format string, v ...any) {
	if config.Debug {
		Logf(format, v...)
	}
}

func Debugln(v ...any) {
	if config.Debug {
		Logln(v...)
	}
}
