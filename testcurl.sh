#!/bin/sh
set -euf
curl --header "Content-Type: applicaton/json" \
	--request POST \
	--data '{"clientKey":"secret398726978clientkey","ip":"192.0.1.222","mac":"x2:01:ab:23:cd:45:ef:67}","uname":"Linux falstaff 4.9.0-8-amd64 #1 SMP Debian 4.9.144-3.1 (2019-02-19) x86_64 GNU/Linux","uptime":" 15:15:22 up 8 days,  2:24, 20 users,  load average: 0.01, 0.04, 0.04"'} \
	http://localhost:9753/api/v1/hello
