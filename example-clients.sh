#!/bin/sh
set -euf

# This shell script read a list of clients from a simpleinventory API server.
# Paul Gorman, 2019

hello="I'm a developer workstation!"

apiServer=localhost
apiPort=9753
readKey=secret1234readkey

# For scripting, "-f" may be better than "-i", so that HTTP errors yield a non-zero unix exit code.
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $readKey" \
	-i \
	http://"$apiServer":"$apiPort"/api/v1/clients
