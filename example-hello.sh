#!/bin/sh
set -euf

# This shell script writes a "hello" check-in to a simpleinventory API server.
# Paul Gorman, 2019

hello="I'm a developer box!"
hostGroup=workstations

apiServer=localhost
apiPort=9753
writeKey=secret1234writekey

ip=$(ip route get $(ip route show | grep default | awk '{ print $3 }') | grep src | awk '{ print $5 }')
machineID=$(cat /etc/machine-id | sha256sum | awk '{ print $1 }')
# Alternate machineID source: $(ip address show up | grep -B 1 "$ip" | head -n 1 | awk '{ print $2 }')
nodeName=$(uname -n)
json=$(printf '{"hello":"%s","hostGroup":"%s","ip":"%s","machineID":"%s","nodeName":"%s"}' \
               "$hello"     "$hostGroup"     "$ip"     "$machineID"     "$nodeName")

# For scripting, "-f" may be better than "-i", so that HTTP errors yield a non-zero unix exit code.
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data "$json" \
	-i \
	http://"$apiServer":"$apiPort"/api/v1/hello
