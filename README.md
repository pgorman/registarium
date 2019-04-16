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
        --data '{"clientKey":"secret398726978clientkey","ip":"192.0.1.222","machineID":"x2:01:ab:23:cd:45:ef:67"}'
        http://localhost:9753/api/v1/hello
```

The only required key is `machineID`.
See `example-client.sh`.

Simple Inventory leaves TLS termination to whatever reverse proxy sits in front of it.

To build, install this dependency (which, itself, relies on `gcc`):

```
$  go get github.com/bvinc/go-sqlite-lite/sqlite3
```


License
----------------------------------------

Copyright 2019 Paul Gorman, and licensed under the 2-clause BSD license.
See `LICENSE.md`.
