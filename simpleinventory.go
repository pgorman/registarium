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

var addr string
var port string
var clientKey string
var charLimit int
var dbFile string

type hello struct {
	ClientKey string `clientKey`
	Hardware  string `json:hardware`
	IP        string `json:ip`
	FirstSeen string `json:firstSeen`
	LastSeen  string `json:lastSeen`
	MAC       string `json:mac`
	NodeName  string `json:nodeName`
	OSRel     string `json:osRel`
	OSSys     string `json:osSys`
	OSVer     string `json:osVer`
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
	db, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.BusyTimeout(5 * time.Second)
	err = db.Exec(`INSERT INTO hellos
		(mac, firstSeen, hardware, ip, lastSeen, nodeName, osRel, osSys, osVer, hello)
		values (?, CURRENT_TIMESTAMP, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
		ip=?, lastSeen=CURRENT_TIMESTAMP, nodeName=?, osRel=?, osSys=?, osVer=?, hello=?`,
		h.MAC, h.Hardware, h.IP, h.NodeName, h.OSRel, h.OSSys, h.OSVer, h.Hello, h.IP, h.NodeName, h.OSRel, h.OSSys, h.OSVer, h.Hello)
	if err != nil {
		log.Println(err)
	}
}

func init() {
	clientKey = os.Getenv("clientKey")
	if len(clientKey) == 0 {
		log.Fatal("please set the 'clientKey' environment variable")
	}

	flag.StringVar(&addr, "address", "127.0.0.1", "network address where we server API")
	flag.StringVar(&port, "port", "9753", "network port to serve API")
	flag.StringVar(&dbFile, "db", "simpleinventory.db", "SQLite database file")
	flag.IntVar(&charLimit, "char-limit", 99, "truncate JSON values supplied by clients at this limit")
	flag.Parse()

	_, err := os.Stat(dbFile)
	if err != nil {
		log.Println("database file", dbFile, "doesn't already exist")
		db, err := sqlite3.Open(dbFile)
		if err != nil {
			log.Println(err)
		}
		defer db.Close()
		db.BusyTimeout(5 * time.Second)
		/*
			Possible sources of values:
			mac= TODO
			hardware=uname -m
			ip (Linux)=ip route get $(ip route show | grep default | awk '{ print $3 }') | grep src | awk '{ print $5 }'
			ip (OpenBSD)=route -n get $(route -n show | grep default | awk '{ print $2 }') | grep 'if address' | awk '{ print $3 }'
			nodeName=uname -n
			osSys=uname -s
			osRel=uname -r
			osVer=uname -v
		*/
		err = db.Exec(`CREATE TABLE IF NOT EXISTS hellos
			(mac TEXT PRIMARY KEY, firstseen NUMERIC, hardware TEXT, ip TEXT,
			lastseen NUMERIC, nodename TEXT, osrel TEXT, ossys TEXT,
			osver TEXT, hello TEXT)`)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	http.HandleFunc("/api/v1/hello", handleHello)
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}
