#!/bin/sh
set -euf

# This shell script writes a "hello" check-in to a simpleinventory API server.
# Paul Gorman, 2019

hello="I'm a developer box!"
hostGroup=workstations

apiServer=localhost
apiPort=9753
writeKey=secret1234writekey

hardware=$(uname -m)
ip=$(ip route get $(ip route show | grep default | awk '{ print $3 }') | grep src | awk '{ print $5 }')
mac=$(ip address show up | grep -B 1 "$ip" | head -n 1 | awk '{ print $2 }')
machineID=$(cat /etc/machine-id)
nodeName=$(uname -n)
osSys=$(uname -s)
osRel=$(uname -r)
osVer=$(uname -v)
json=$(printf '{"hardware":"%s","hostGroup":"%s","ip":"%s","mac":"%s","machineID":"%s","nodeName":"%s","osRel":"%s","osSys":"%s","osVer":"%s","hello":"%s"}' \
               "$hardware"      "$hostGroup"     "$ip"     "$mac"     "$machineID"     "$nodeName"     "$osRel"     "$osSys"     "$osVer"     "$hello")

# For scripting, "-f" may be better than "-i", so that HTTP errors yield a non-zero unix exit code.
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data "$json" \
	-i \
	http://"$apiServer":"$apiPort"/api/v1/hello
