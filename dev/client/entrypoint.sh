#!/bin/bash
set -euo pipefail

printf 'register\n' | /usr/bin/larsys -host-ip=larsys-server -host-port=5454 -log=/var/log/larsys/daemon.log
printf 'register\n' | /usr/bin/larsys -host-ip=larsys-server -host-port=5454 -log=/var/log/larsys/daemon.log
printf 'revoke\n' | /usr/bin/larsys -host-ip=larsys-server -host-port=5454 -log=/var/log/larsys/daemon.log
printf 'revoke\n' | /usr/bin/larsys -host-ip=larsys-server -host-port=5454 -log=/var/log/larsys/daemon.log
printf 'register\n' | /usr/bin/larsys -host-ip=larsys-server -host-port=5454 -log=/var/log/larsys/daemon.log
