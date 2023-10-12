v0.1.0
======

Release date: 2023-10-09

Hello world.

This marks the initial experimental and incomplete version of Svipul.
Considerable changes are expected to every aspect, but we need to start
releases of some sort sooner or later.

Here are some feature hilights:

- Connects to a RabbitMQ queue, with hardcoded URL and credentials
- Reads "orders" from the queue
- Can poll SNMP based on orders, supports GET and (bulk) WALK
- Can build table maps, cache them, and build GET requests based on element
  lists: E.g.: You can poll a router for ifHCInOctets and ifHCOutOctets on
  every interface whos ifName matches the regex (ge|xt|et)-.*
- Results are transmitted using Skogul (as a library). Anything Skogul can
  do, Svipul can do (except no receivers). Uses Skogul v0.25.1.
- Bundles standard mibs, supports others. Uses gosmi to do various lookups
  and renderings.

There are two binaries:

- ``worker`` is the actual Svipul SNMP worker
- ``addjob`` is a tiny simple tool to add jobs to the queue

There are more features not listed here. Probably.

Some things that are MISSING:

- Configurable AMQP url/credentials
- Proper release-quality builds (e.g.: RPM, config directory, etc)
- A ton of test cases
- A stable API

There are also a lot of components in the Svipul "eco-system" that are
missing. They may or may not become part of this repository in the future.
