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
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var clientKey string
var dbFile string

type hello struct {
	ClientKey string `clientKey`
	IP        string `json:ip`
	LastSeen  string `json:lastSeen`
	MAC       string `json:mac`
	Uname     string `json:uname`  // Output of `uname -a`
	Uptime    string `json:uptime` // Output of `uptime`
	Hello     string `json:hello`
}

// hearHello decodes a "hello" POST from a client.
func handleHello(w http.ResponseWriter, r *http.Request) {
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
	conn, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	// It's always a good idea to set a busy timeout
	conn.BusyTimeout(5 * time.Second)
}

func main() {
	clientKey = os.Getenv("clientKey")
	if len(clientKey) == 0 {
		log.Fatal("please set the 'clientKey' environment variable")
	}
	var addr, port string
	flag.StringVar(&addr, "address", "127.0.0.1", "network address where we server API")
	flag.StringVar(&port, "port", "9753", "network port to serve API")
	flag.StringVar(&dbFile, "db", "simpleinventory.db", "SQLite database file")
	flag.Parse()
	_, err := os.Stat(dbFile)
	if err != nil {
		log.Println("database file", dbFile, "doesn't already exist")
	}
	http.HandleFunc("/api/v1/hello", handleHello)
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}
