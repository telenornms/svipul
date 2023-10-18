===========
svipul-snmp
===========

-----------
Svipul SNMP
-----------

:Manual section: 1
:Authors: Kristian Lyngstøl
:Date: 18.10.2023
:Version: 0.1.0-dirty

SYNOPSIS
========

::

        svipul-snmp [-broker string] [-debug] [-workers int]

DESCRIPTION
===========

Svipul is a toolset for collecting data from network devices. svipul-snmp
details the SNMP collector/poller.

Svipul listens to instructions, called orders, on a RabbitMQ queue and
transmits results back.

It is in heavy development. Expect significant changes.



OPTIONS
=======

-f string
        configuration file to read (default: "/etc/svipul/snmp.toml")

-debug
  	enable debug

SEE ALSO
========

* svipul-addjob(1)

BUGS
====

Yes.

See https://github.com/telenornms/svipul for more.

COPYRIGHT
=========

This document is licensed under the same license as Svipul itself. See
LICENSE for details.

* Copyright 2023 Telenor Norge AS
