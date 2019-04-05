// Copyright 2019 Paul Gorman.
// Licensed under the 2-Clause BSD License.
// See `LICENSE.md`.

// `simpleinventory` lets client self-report their whereabouts on the network.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
)

var clientKey string

type hello struct {
	ClientKey string `clientKey`
	IP        string `json:ip`
	LastSeen  string `json:lastSeen`
	MAC       string `json:mac`
	Uname     string `json:uname`  // Output of `uname -a`
	Uptime    string `json:uptime` // Output of `uptime`
}

// hearHello decodes a "hello" POST from a client.
func hearHello(w http.ResponseWriter, r *http.Request) {
	var h hello
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&h)
	if err != nil {
		log.Println(err)
	}
	if h.ClientKey != clientKey {
		log.Println("client sent wrong API key")
		return
	}
	log.Println(h)
}

func main() {
	clientKey = os.Getenv("clientKey")
	if len(clientKey) == 0 {
		log.Fatal("please set the 'clientKey' environment variable")
	}
	var addr, port string
	flag.StringVar(&addr, "a", "127.0.0.1", "network address where we server API")
	flag.StringVar(&port, "p", "9753", "network port to serve API")
	flag.Parse()
	http.HandleFunc("/api/v1/hello", hearHello)
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}
