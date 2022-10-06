#!/bin/sh

ORIGPWD=$PWD
TMP=$(mktemp -d)
set -x
set -e
cd $TMP
wget ftp://ftp.cisco.com/pub/mibs/v2/v2.tar.gz
wget https://www.juniper.net/techpubs/software/junos/junos151/juniper-mibs-15.1R3.6.tgz
tar xvzf v2.tar.gz  --strip-components=2
tar xvzf juniper-mibs-15.1R3.6.tgz
mkdir -p mibs

mv v2 mibs/CiscoMibs
mv StandardMibs JuniperMibs mibs/
mkdir mibs/modules
set +x
echo "Copying mibs to mids/modules/(MODULE NAME)"
for a in mibs/CiscoMibs/* mibs/StandardMibs/* mibs/JuniperMibs/*; do n=$(awk '/DEFINITION.*BEGIN/ { print $1; exit 0; }' $a); cp $a mibs/modules/$n || echo "failed: $a" ; done
mv mibs ${ORIGPWD}/
cd ${ORIGPWD}
rm -rf ${TMP}
