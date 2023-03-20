===================================
Various extras/dev stuff for Svipul
===================================

Testing snmp
============

Install snmpd (``apt install snmpd``).

Expose something - the snmpd.conf supplied is extremely open and easy to
test. It uses the "public" community (community is the authentication
token/word/"pw" in this context) and exposes "everything". Don't use this
on something exposed to the internet, but local is fine.

Remember to (re)start the snmpd service.

Many of MY tests use fake /etc/hosts. Typically::

	# cat /etc/hosts

	192.168.122.52	lol
	192.168.122.41	vm-lol1 vm-lol2 vm-lol3 vm-lol4 vm-lol5
	192.168.2.3 ex-lol1 ex-lol2 ex-lol3 ex-lol4 ex-lol5 ex-lol6 ex-lol7 ex-lol8 ex-lol9 ex-lol10 ex-lol11 ex-lol12 ex-lol13 ex-lol14 ex-lol15 ex-lol16 ex-lol17 ex-lol18 ex-lol19 ex-lol20 ex-lol21 ex-lol22 ex-lol23 ex-lol24 ex-lol25 ex-lol26 ex-lol27 ex-lol28 ex-lol29 ex-lol30 ex-lol31 ex-lol32 ex-lol33 ex-lol34 ex-lol35 ex-lol36 ex-lol37 ex-lol38 ex-lol39 ex-lol40 ex-lol41 ex-lol42 ex-lol43 ex-lol44 ex-lol45 ex-lol46 ex-lol47 ex-lol48 ex-lol49 ex-lol50 

The reason for duplicates is to fake the pressence of multiple devices for
the sake of lock-testing/avoidance. It is obviously not required - but
since many of the example-orders use these host names, they might help you
(I'm very good at naming stuff).

To verify, you can use ``snmpwalk`` command (part of the snmp package), two
examples::

	# snmpwalk -v2c -mALL -c public vm-lol2

	# snmpwalk -v2c -c public vm-lol2
