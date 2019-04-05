Simple Inventory
========================================

Imagine we have a bunch of Linux boxes.
We want to know where they are on our network.
We're content for them to self-report.

This simple JSON API server saves client reports to SQLite.

The clients can report to the server as simply as:

```
curl --header "Content-Type: applicaton/json" \
        --request POST \
        --data '{"clientKey":"secret398726978clientkey","ip":"192.0.1.222","mac":"x2:01:ab:23:cd:45:ef:67}","uname":"Linux falstaff 4.9.0-8-amd64 #1 SMP Debian 4.9.144-3.1 (2019-02-19) x86_64 GNU/Linux","uptime":" 15:15:22 up 8 days,  2:24, 20 users,  load average: 0.01, 0.04, 0.04"'} \
        http://localhost:9753/api/v1/hello
```

Simple Inventory leaves TLS terminaltion to whatever reverse proxy sits in front of it.

To build, install this dependency (which, itself, relies on `gcc`):

```
$  go get github.com/bvinc/go-sqlite-lite/sqlite3
```


License
----------------------------------------

Copyright 2019 Paul Gorman, and licensed under the 2-clause BSD license.
See `LICENSE.md`.
