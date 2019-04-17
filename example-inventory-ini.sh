#!/bin/sh
set -euf

# This shell script pulls an Ansible-compatible inventory list form the API server.
# Paul Gorman, 2019

apiServer=localhost
apiPort=9753
readKey=secret1234readkey

curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $readKey" \
	-f \
	http://"$apiServer":"$apiPort"/api/v1/inventory
