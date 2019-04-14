#!/bin/sh
set -euf

clientKey=secret398726978clientkey
ip=$(ip route get $(ip route show | grep default | awk '{ print $3 }') | grep src | awk '{ print $5 }')
mac=$(ip address show up | grep -B 1 "$ip" | head -n 1 | awk '{ print $2 }')
nodeName=$(uname -n)
osSys=$(uname -s)
osRel=$(uname -r)
osVer=$(uname -v)
hello="I'm a thin client!"
json=$(printf '{"clientKey":"%s","ip":"%s","mac":"%s", "hello":"%s"}' "$clientKey" "$ip" "$mac" "$hello")

curl --header "Content-Type: applicaton/json" \
	--request POST \
	--data "$json" \
	http://localhost:9753/api/v1/hello
