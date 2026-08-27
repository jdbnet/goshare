#!/bin/bash
id goshare &>/dev/null || useradd -r -s /usr/sbin/nologin goshare
mkdir -p /var/lib/goshare/data
chown -R goshare:goshare /var/lib/goshare
