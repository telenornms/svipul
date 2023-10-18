=============
svipul-addjob
=============

-------------
Svipul addjob
-------------

:Manual section: 1
:Authors: Kristian Lyngstøl
:Date: 18.10.2023
:Version: 0.1.0-dirty

SYNOPSIS
========

::

        svipul-addjob [-broker string] [-debug] [-workers int]

DESCRIPTION
===========

Svipul is a toolset for collecting data from network devices. svipul-addjob
is a small utility for adding jobs to the rabbitMQ queue.

It is in heavy development. Expect significant changes.



OPTIONS
=======

-broker string
  	AMQP broker-url to connect to (default "amqp://guest:guest@localhost:5672/")

-delay duration
  	delay between individual orders, negative value means only one execution (default -1s)

-sleep duration
  	sleep between iterations, negative value means only one execution (default -1s)

-ttl duration
  	expiry time. Minimum: 1ms (default 30s)

SEE ALSO
========

* svipul-snmp(1)

BUGS
====

Yes.

See https://github.com/telenornms/svipul for more.

COPYRIGHT
=========

This document is licensed under the same license as Svipul itself. See
LICENSE for details.

* Copyright 2023 Telenor Norge AS
