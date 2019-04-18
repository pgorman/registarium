#!/bin/sh
set -euf

# This shell script populates sample data for the inventory API server.
# Paul Gorman, 2019

apiServer=localhost
apiPort=9753
writeKey=secret1234writekey

curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"2d55ee0875b046fbb57ac830f262c4b3","ip":"192.0.2.32","nodeName":"alice","hostGroup":"workstation"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"4d55ee0875b046ffb07ac830f262c441","ip":"192.0.2.112","nodeName":"bob","hostGroup":"workstation"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"b155ee0875b046ffb07ac530f269cc4c","ip":"192.0.2.209","nodeName":"charlie","hostGroup":"workstation"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"a653ed0875b046ffb07ac530f269cca4","ip":"192.0.2.20","nodeName":"mars","hostGroup":"hypervisor"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"9b53ed0875b046ffb07ac530f2896a22","ip":"192.0.2.21","nodeName":"venus","hostGroup":"hypervisor"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"7253ef0845b046ffb07ac530f289c162","ip":"192.0.2.39","nodeName":"arrow","hostGroup":"smarthost"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"2293ef0845b046ffb07ac5b0ff89cd51","ip":"192.0.2.36","nodeName":"dart","hostGroup":"smarthost"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"b233ef0845b046ffb07ac5b0ff89c45a","ip":"203.0.113.181","nodeName":"tc0034","hostGroup":"thinclient"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"1273ef0845b046ffb07ac5b0ff89ca43","ip":"203.0.113.213","nodeName":"tc2437","hostGroup":"thinclient"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"cc731f0845b046ffb07ac5b0ff6f2b1b","ip":"203.0.113.120","nodeName":"tc1132","hostGroup":"thinclient"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
curl --header "Content-Type: applicaton/json" \
	--header "Authorization: ApiKey $writeKey" \
	--data '{"machineID":"22d31f0845b046ffb07ac5b0f6bb7baa","ip":"203.0.113.144","nodeName":"tc1488","hostGroup":"thinclient"}' \
	http://"$apiServer":"$apiPort"/api/v1/hello
