Simple Inventory
========================================

Imagine we have a bunch of Linux boxes.
We want to know where they are on our network.
We're content for them to self-report.

This simple JSON API server saves client reports to SQLite.

The clients can report to the server as simply as:

```
curl --header "Content-Type: application/json" \
	--request POST \
	--data '{"apiKey":"mysecretkey123",mac":"x2:ab:01:cd:23:ef","ipAddr":"10.0.99.123","hello":"I am a Raspberry Pi 3 bullentin board display."}' \
	https://inventory.example.com/api/v1/hello
```


License
----------------------------------------

Copyright 2019 Paul Gorman, and licensed under the 2-clause BSD license.
See `LICENSE.md`.
