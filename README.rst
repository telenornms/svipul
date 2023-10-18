Svipul
======

Svipul is a set of tools for gathering data from network devices. Today,
only through SNMP. In the future, the architecture should support multiple
protocols for collection.

The design goals of Svipul are, roughly:

- Scale horizontally, support massive scaling
- Shield network hardware from accidental or intentional flooding
- Theoretically replace all use cases of SNMP data collection
- Stateless worker(s), no concept of a set list of machines/boxes (e.g.:
  dynamically configured targets)
- Modularized architecture, no tight coupling.

This repository contains the SNMP worker and a simple convenience tool for
manually adding orders. The first parts written.

The worker relies on RabbitMQ to queue orders for individual polling-jobs.

The SNMP worker uses Skogul (https://github.com/telenornms/skogul) to
transmit the results. Skogul supports a large number of different protocols
and transformations, making Svipul storage-agnostic.

Testing
-------

Install go, make and RabbitMQ.

Run ``make`` (or ``make install``).

Start rabbitmq, and ``./svipul-snmp``.

Test stuff with ``./svipul-addjob docs/examples/orders/...`` (or make your
own order).

The default binary uses localhost and guest:guest for authentication of
rabbitmq, but this can be configured.

To do something reasonable with the output, you can configure the
skogul-part. See skogul's documentation for details. Svipul works by using
the skogul handler named "svipul", and otherwise uses everything you can do
with Skogul (except set up other receivers, because why would you?).

Current feature-set and stability
---------------------------------

Most of the work so far has gone into making the SNMP worker stable, fast
and support a set of key features. All other components are, at the moment,
non-existent. But because Svipul sends the results with Skogul and can take
orders, it is still fully possible to use it for SNMP polling.

The worker currently has the following basic features:

- A single worker will only ever perform a single request for a given
  target at a time: If multiple orders to poll a target is received at the
  same time, only one will be carried out on a worker. In the future, this
  exclusivity will be extended to include all workers.
- Fetch data using GET or BULK WALK
- Parse MIB files (using gosmi). Standard mibs are bundled. Others can be
  configured.
- SMI: You can request data by numeric OID or logical name, or a mix.
- Optimized "table gets"/GET element: If fetching data from a table, e.g.
  ifTable, Svipul can build a name<->index cache and fetch each counter
  with precise GET, even using regular expressions to match the elements in
  the table, e.g.: fetch ifHCInOctets/ifHCOutOctets for interfaces matching
  ge-.* and et-.*.
- Stateless, horizontally scaled

Stability and performance
-------------------------

Svipul is still considered experimental, but is slowly being put into
limited production.

We do not currently have hard data on how fast it is, or how stable. But
tests indicate that it is fast.

But if you are considering using it in production, you should probably
expect some surprises. Through the entire 0.1.x release, basically
everything can change.

Basic design
------------

Design is subject to change, massively.

.. image:: docs/pollng.drawio.png

The basic concept is to have independent pollers that listen for tasks on a
queue. A locking service is needed too, but not (currently) shown. The
locking service is to prevent independent pollers from working on the same
device.

Each poller then uses Skogul to report back. This initial implementation
has Skogul embeded, thus an external one is not strictly speaking needed.

Results are streamed to a pub/sub-style exchange on RabbitMQ, and multiple
listeners can wait for it.

Each component is meant to be largely independent, and the design is meant
as a reference design where one need not implement all of it for it to be
useful. E.g.: You can get away with a simple poller, a shell script for
scheduler and skogul writing to influxDB directly if need be, using what
already exists today.

Name
----

Svipul is a Valkyrie. Valkyries are cool. Working name until recently was
tpoll, but it just changed.
