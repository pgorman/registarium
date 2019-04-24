Registarium
========================================

Registarium, a Go API server, saves client registrations to SQLite, and produces an Ansible-compatible inventory file from that data.

Imagine we have a bunch of Linux boxes.
We want to know where they are on our network.
We're content for them to self-report.

The clients can report to the server as simply as:

```
curl --header "Content-Type: applicaton/json" \
	--request POST \
	--header "Authorization: ApiKey 123mysupersecretkey" \
	--data '{"hostGroup":"workstations","ip":"192.0.1.222","machineID":"x201ab23cd45ef67"}'
	http://localhost:9753/api/v1/hello
```

The only required data is `machineID`.
See `example-client.sh`.

Registarium leaves TLS termination to whatever reverse proxy sits in front of it.
When using a reverse proxy, remember to set the client's original IP in a header like `Forwarded` or `X-Real-IP` to keep logging and debug messages helpful.


Getting Started
----------------------------------------

If you have not already set up a Go build environment, follow the [Go getting started instructions](https://golang.org/doc/install).

Get the Registarium source and its one dependency:

```
$  cd $GOPATH/src
$  go get github.com/pgorman/registarium
$  go get github.com/bvinc/go-sqlite-lite/sqlite3

```

Note that the sqlite3 package uses CGO, so it needs a minimal C toolchain.
On Debian-based systems, this should be sufficient to pull in `gcc` and `libc6-dev`:

```
#  apt install build-essential
```

Build and test Registarium:

```
$  cd $GOPATH/src/github.com/pgorman/registarium
$  go build
$  readKey=secret1234readkey writeKey=secret1234writekey ./registarium --debug
$  ./populate-test-data.sh
$  ./example-inventory-ini.sh
```


Deploying Registarium
----------------------------------------

Build a binary for your target deployment platform.
If that matches your build workstation, `go build` should be enough to produce a `registarium` binary.
If, for example, you're building on amd64 for deployment on 386, cross-compile like:

```
$  cd $GOPATH/src/github.com/pgorman/registarium
$  CGO_ENABLED=1 GOOS=linux GOARCH=386 go build
$  scp ./registarium myserver:
```

If using systemd to supervise Registarium, customize the API keys in `registarium.service` and copy that file to the deployment server too.

Decide where to install the binary and save the inventory data.
For example:

```
myserver#  chown root:staff registarium
myserver#  chmod 555 registarium
myserver#  mv registarium /usr/local/bin/
myserver#  chown root:root registarium.service
myserver#  chmod 600 registarium.service
myserver#  mv registarium.service /etc/systemd/system/
myserver#  sudo mkdir -p /var/local/registarium
myserver#  chown root:staff /var/local/registarium
myserver#  systemctl enable registarium.service
myserver#  systemctl start registarium.service
```

Finally, configure your reverse proxy (e.g., HAProxy, Nginx, Apache) to do TLS termination and proxying to Rregistarium.
A reverse proxy configuration for Apache, with the `proxy` and `proxy_http` modules loaded, looks something like:

```
<VirtualHost *:443>
	ServerName inventory.example.com
	SSLEngine ON
	SSLVerifyClient optional
	SSLCertificateKeyFile /etc/ssl/private/STAR_example_com.key
	SSLCertificateFile /etc/ssl/certs/STAR_example_com.crt
	SSLCertificateChainFile /etc/ssl/certs/STAR_example_com.ca-bundle
	ProxyPreserveHost On
	ProxyPass "/"  "http://127.0.0.1:9753/"
	ProxyPassReverse "/"  "http://127.0.0.1:9753/"
</VirtualHost>
```


Links
----------------------------------------

- https://docs.ansible.com/ansible/2.6/dev_guide/developing_inventory.html


License
----------------------------------------

Copyright 2019 Paul Gorman, and licensed under the 2-clause BSD license.
See `LICENSE.md`.
