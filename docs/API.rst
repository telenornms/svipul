Svipul API
==========

The Svipul API works by accepting orders over a RabbitMQ queue, and sending
the reply using Skogul (as a library).


``svipul-api``
--------------

Most user-interaction is probably going through an external API, one such
is the currently proprietary Svipul-API, from now written as ``svipul-api``.
The ``svipul-api`` offers a HTTP API to add orders.

While ``svipul-api`` is a separate code base, the only practical difference
for end-users is that it allows you to post multiple orders as a single
array of JSON objects.

E.g., the core API of Svipul would accept an order over RabbitMQ looking
like::

        {
                "target": "vm-lol1",
                "id": "somethingRandomForEachRequest",
                "mode": "Get",
                "oids": [ ".1.3.6.1.2.1.1.5.0", ".1.3.6.1.2.1.1.1.0"]
        }

To post multiple orders, you need to post multiple messages with one order
per message to RabbitMQ.

The exact same order over ``svipul-api`` would look like::

        [
                {
                        "target": "vm-lol1",
                        "id": "somethingRandomForEachRequest",
                        "mode": "Get",
                        "oids": [ ".1.3.6.1.2.1.1.5.0", ".1.3.6.1.2.1.1.1.0"]
                }
        ]

To post multiple orders through ``svipul-api``, you could either post
multiple HTTP requests with one order, or a single HTTP request with an
array of orders. In the latter case ``svipul-api`` will take
responsibility for splitting it up into individual messages.

The remainder of this API documentation is written as if ``svipul-api``
does not exist, and deals only with the individual orders.

Basic order
-----------

The core mechanic of Svipul revolve around orders. An order is an
instruction to do some work related to a single host. Orders are ephemeral
and if you want a device polled every Nth minute, you need to post an order
every Nth minute.

Svipul will avoid sending multiple requests to the same device at the same
time. Failed requests are retried exactly once at a randomized delay,
between 1 and 10 seconds later. There is no guarantee that the request will
succeed after this. Future version may provide better error reporting, but
current version do not.

An order is expressed as a JSON object.

All possible fields in an order are::

	Target    string   // Host/target
	Oids      []string // OIDs, also accepts logical names (e.g.: ifName)
	Elements  []string // Elemnts, if GetElements mode. Elements == interfaces (could be other in the future)
	Key       string   // Map key to use for looking up elements
	Mode      Mode     // What mode to use
	Community string   `json:",omitempty"` // Community to use, blank == figure it out yourself/use default (meaning depends on issuer)
	ID        string   `json:",omitempty"`
	Result    ResolveM // Auto (default) = resolve based on input, OID = leave OIDs unresolved, Resolve = try to resolve

Target, Oids and Community is considered sufficiently explained above.

ID is reflected back into the metadata of the result and has no other
function than to allow a caller to identify the result of its request.

The Mode defines how Svipul will carry out this specific order. Not all
mode requires/uses all the other fields in an order. The possible modes
are::

	Walk                    // Do a walk
	Get                     // Get just these oids
	GetElements             // Get these specific oids, but per elements
	BuildMap                // Build an OMap
	ClearMap                // Clear the OMap cache

The JSON representation is case in-sensitive.

Output
------

The result/output of Svipul is sent using Skogul, which means there is a
large degree of variety. In theory, the result could be stored to files
using GOB encoding, sent over HTTPS using Influx-encoding or any number of
methods. Describing all possible output-formats is beyond the scope of this
document, see Skogul for the full documentation.

The documentation here provides examples of un-transformed JSON output. The
most common transformation applied will be splitting nested data of, e.g.,
GetElements into individual metrics.

**However**, the primary use-case so far is storing data in InfluxDB. If
this applies to you, assume that anything referred to as Metadata is
available as tags, and the rest is data.

GET
---

Parameters used: `Target`, `Oids`, `Mode`, `Community`, `Result`, `ID`

Typically used to fetch data that is not part of a table. E.g.:
``sysName``, ``sysUptime``, etc.

Example::

        {
                "target": "vm-lol1",
                "id": "somethingRandomForEachRequest",
                "mode": "Get",
                "oids": [ "sysName.0"]
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T15:22:57.092894205+02:00",
              "metadata": {
                "id": "somethingRandomForEachRequest",
                "target": "vm-lol1"
              },
              "data": {
                "0": {
                  "sysName": "debian11"
                }
              }
            }
          ]
        }

This polls `vm-lol1` for ``sysName.0``. Fully numeric OIDs can also be
used::

        {
                "target": "vm-lol1",
                "id": "somethingRandomForEachRequest",
                "mode": "Get",
                "oids": [ ".1.3.6.1.2.1.1.5.0", ".1.3.6.1.2.1.1.1.0"]
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T15:23:37.092673394+02:00",
              "metadata": {
                "id": "somethingRandomForEachRequest",
                "target": "vm-lol1"
              },
              "data": {
                ".1.3.6.1.2.1.1.1.0": "Linux debian11 5.10.0-23-amd64 #1 SMP Debian 5.10.179-3 (2023-07-27) x86_64",
                ".1.3.6.1.2.1.1.5.0": "debian11"
              }
            }
          ]
        }

If ``Result`` is not set, the result will be provided to best match the
input, though this is not guaranteed if a mix of numeric and symbolic OIDs
are used.

This behavior can be overridden. If you want numeric IDs, set ``"Result":
"OID"``, if you want symbolic names, set ``"Result": "Resolve"``. Example::

        {
                "target": "vm-lol1",
                "id": "somethingRandomForEachRequest",
                "mode": "Get",
                "oids": [ ".1.3.6.1.2.1.1.5.0", ".1.3.6.1.2.1.1.1.0"],
                "result": "Resolve"
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T15:26:43.892432176+02:00",
              "metadata": {
                "id": "somethingRandomForEachRequest",
                "target": "vm-lol1"
              },
              "data": {
                "0": {
                  "sysDescr": "Linux debian11 5.10.0-23-amd64 #1 SMP Debian 5.10.179-3 (2023-07-27) x86_64",
                  "sysName": "debian11"
                }
              }
            }
          ]
        }

Walk
----

Parameters used: `Target`, `Oids`, `Mode`, `Community`, `Result`, `ID`

Does a bulkwalk, which is generally slow. Typical use case: Discovery,
one-offs.

Example::

        {
                "target": "vm-lol1",
                "mode": "Walk",
                "oids": ["ifTable", "ifXTable"]
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T15:28:26.592024002+02:00",
              "metadata": {
                "target": "vm-lol1"
              },
              "data": {
                "1": {
                  "ifAdminStatus": "up(1)",
                  "ifAlias": "",
                  "ifConnectorPresent": "false(2)",
                  "ifCounterDiscontinuityTime": 0,
                  "ifDescr": "lo",
                  "ifHCInBroadcastPkts": 0,
                  "ifHCInMulticastPkts": 0,
                  "ifHCInOctets": 2990,
                  "ifHCInUcastPkts": 30,
                  "ifHCOutBroadcastPkts": 0,
                  "ifHCOutMulticastPkts": 0,
                  "ifHCOutOctets": 2990,
                  "ifHCOutUcastPkts": 30,
                  "ifHighSpeed": 10,
                  "ifInBroadcastPkts": 0,
                  "ifInDiscards": 0,
                  "ifInErrors": 0,
                  "ifInMulticastPkts": 0,
                  "ifInNUcastPkts": 0,
                  "ifInOctets": 2990,
                  "ifInUcastPkts": 30,
                  "ifInUnknownProtos": 0,
                  "ifIndex": 1,
                  "ifLastChange": 0,
                  "ifMtu": 65536,
                  "ifName": "lo",
                  "ifOperStatus": "up(1)",
                  "ifOutBroadcastPkts": 0,
                  "ifOutDiscards": 0,
                  "ifOutErrors": 0,
                  "ifOutMulticastPkts": 0,
                  "ifOutNUcastPkts": 0,
                  "ifOutOctets": 2990,
                  "ifOutQLen": 0,
                  "ifOutUcastPkts": 30,
                  "ifPhysAddress": "",
                  "ifPromiscuousMode": "false(2)",
                  "ifSpecific": "",
                  "ifSpeed": 10000000,
                  "ifType": "softwareLoopback(24)"
                },
                "2": {
                  "ifAdminStatus": "up(1)",
                  "ifAlias": "",
                  "ifConnectorPresent": "true(1)",
                  "ifCounterDiscontinuityTime": 0,
                  "ifDescr": "Red Hat, Inc. Device 0001",
                  "ifHCInBroadcastPkts": 0,
                  "ifHCInMulticastPkts": 0,
                  "ifHCInOctets": 89793731,
                  "ifHCInUcastPkts": 52464,
                  "ifHCOutBroadcastPkts": 0,
                  "ifHCOutMulticastPkts": 0,
                  "ifHCOutOctets": 2841057,
                  "ifHCOutUcastPkts": 18283,
                  "ifHighSpeed": 0,
                  "ifInBroadcastPkts": 0,
                  "ifInDiscards": 9354,
                  "ifInErrors": 0,
                  "ifInMulticastPkts": 0,
                  "ifInNUcastPkts": 0,
                  "ifInOctets": 89793731,
                  "ifInUcastPkts": 52464,
                  "ifInUnknownProtos": 0,
                  "ifIndex": 2,
                  "ifLastChange": 305,
                  "ifMtu": 1500,
                  "ifName": "enp1s0",
                  "ifOperStatus": "up(1)",
                  "ifOutBroadcastPkts": 0,
                  "ifOutDiscards": 0,
                  "ifOutErrors": 0,
                  "ifOutMulticastPkts": 0,
                  "ifOutNUcastPkts": 0,
                  "ifOutOctets": 2841057,
                  "ifOutQLen": 0,
                  "ifOutUcastPkts": 18283,
                  "ifPhysAddress": "52:54:00:cb:64:19",
                  "ifPromiscuousMode": "false(2)",
                  "ifSpecific": "",
                  "ifSpeed": 0,
                  "ifType": "ethernetCsmacd(6)"
                }
              }
            }
          ]
        }

While this is a tempting thing to use, be aware that walk is generally very
slow on network hardware.

GetElements
-----------

Parameters used: `Target`, `Oids`, `Mode`, `Community`, `Result`, `ID`,
`Key`, `Elements`

A work horse, used to fetch data from a table using GET (BULK GET). It will
build and cache a map of elements to indexes based on a key behind the
scenes. If no key is provided "ifName" is assumed.

Example::

        {
                "target": "vm-lol2",
                "mode": "GetElements",
                "oids": ["ifHighSpeed", "ifType", "ifName", "ifDescr", "ifAlias", "ifOperStatus", "ifAdminStatus", "ifLastChange", "ifPhysAddress", "ifHCInOctets", "ifHCOutOctets", "ifInDiscards", "ifOutDiscards", "ifInErrors", "ifOutErrors", "ifInUnknownProtos", "ifOutQLen" ],
                "elements": ["e.*"]
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T16:22:33.192380088+02:00",
              "metadata": {
                "target": "vm-lol2"
              },
              "data": {
                "enp1s0": {
                  "ifAdminStatus": "up(1)",
                  "ifAlias": "",
                  "ifDescr": "Red Hat, Inc. Device 0001",
                  "ifHCInOctets": 102571068,
                  "ifHCOutOctets": 3707596,
                  "ifHighSpeed": 0,
                  "ifInDiscards": 11085,
                  "ifInErrors": 0,
                  "ifInUnknownProtos": 0,
                  "ifLastChange": 305,
                  "ifName": "enp1s0",
                  "ifOperStatus": "up(1)",
                  "ifOutDiscards": 0,
                  "ifOutErrors": 0,
                  "ifOutQLen": 0,
                  "ifPhysAddress": "52:54:00:cb:64:19",
                  "ifType": "ethernetCsmacd(6)"
                }
              }
            }
          ]
        }

Behind the scenes, Svipul will first do a Walk for ``ifName`` and map index
to ``ifName``. This is cached. Then, a regular expression match for ``e.*``
is executed to build precise requests for each OID requested. In this
example, only a single device matched.

An other example, using ``hrSWInstalledTable``::

        {
                "target": "vm-lol2",
                "mode": "GetElements",
                "oids": [ "hrSWInstalledIndex", "hrSWInstalledName", "hrSWInstalledID","hrSWInstalledType", "hrSWInstalledDate" ],
                "ooids": [ "hrSWInstalledName", "hrSWInstalledDate" ],
                "key": "hrSWInstalledName",
                "elements": [".*chrom.*"]
        }

Result::

        {
          "metrics": [
            {
              "timestamp": "2023-10-09T16:24:57.193007456+02:00",
              "metadata": {
                "target": "vm-lol2"
              },
              "data": {
                "chromium-common_115.0.5790.170-1~deb11u1_amd64": {
                  "hrSWInstalledDate": "2023-8-7,15:30:56.0,+2:0",
                  "hrSWInstalledID": "",
                  "hrSWInstalledIndex": 33,
                  "hrSWInstalledName": "chromium-common_115.0.5790.170-1~deb11u1_amd64",
                  "hrSWInstalledType": "application(4)"
                },
                "chromium-sandbox_115.0.5790.170-1~deb11u1_amd64": {
                  "hrSWInstalledDate": "2023-8-7,15:30:56.0,+2:0",
                  "hrSWInstalledID": "",
                  "hrSWInstalledIndex": 34,
                  "hrSWInstalledName": "chromium-sandbox_115.0.5790.170-1~deb11u1_amd64",
                  "hrSWInstalledType": "application(4)"
                },
                "chromium_115.0.5790.170-1~deb11u1_amd64": {
                  "hrSWInstalledDate": "2023-8-7,15:30:56.0,+2:0",
                  "hrSWInstalledID": "",
                  "hrSWInstalledIndex": 32,
                  "hrSWInstalledName": "chromium_115.0.5790.170-1~deb11u1_amd64",
                  "hrSWInstalledType": "application(4)"
                },
                "libchromaprint1_1.5.0-2_amd64": {
                  "hrSWInstalledDate": "2022-3-2,13:11:59.0,+1:0",
                  "hrSWInstalledID": "",
                  "hrSWInstalledIndex": 339,
                  "hrSWInstalledName": "libchromaprint1_1.5.0-2_amd64",
                  "hrSWInstalledType": "application(4)"
                }
              }
            }
          ]
        }

In this case, the first request took 90ms, because it built a cache first
(and a Linux VM is fast). The table is quite large, with several thousand
entries. But subsequent requests took 20-30ms. This speed-up is far more
prominent in network hardware.

BuildMap
--------

Parameters used: `Target`, `Community`, `Key`

BuildMap explicitly builds and caches a map for later use, e.g., ifName
<->index.

Example::

        {
                "target": "vm-lol2",
                "mode": "BuildMap",
                "key": "ifName"
        }

There is no result, but the map is cached, to be used by GetElements later.

Note that this is NOT required - if no map exists when GetElements is
called, it will be built (and cached) on demand!

FIXME: Future versions will include explicit TTL support. For now, things
are cached for an hour, configurable on startup.

ClearMap
--------

Parameters used: `Target`, `Community`, `Key`

Explicitly clears the map cache for a target/key combination. The opposite
of what BuildMap does.

Example::

        {
                "target": "vm-lol2",
                "mode": "ClearMap",
                "key": "ifName"
        }

Not required for regular use since things will automatically time out, and
issuing BuildMap will always update the cache.

