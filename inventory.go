// Copyright 2019 Paul Gorman.
// Licensed under the 2-Clause BSD License.
// See `LICENSE.md`.

// `inventory` lets client self-report their whereabouts on the network.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var addr string
var readKey string
var writeKey string
var charLimit int
var dbFile string
var debug bool
var port string

type client struct {
	FirstSeen string `json:"firstSeen"`
	Hardware  string `json:"hardware"`
	HostGroup string `json:"hostGroup"`
	IP        string `json:"ip"`
	LastSeen  string `json:"lastSeen"`
	MAC       string `json:"mac"`
	MachineID string `json:"machineID"`
	NodeName  string `json:"nodeName"`
	OSRel     string `json:"osRel"`
	OSSys     string `json:"osSys"`
	OSVer     string `json:"osVer"`
	Hello     string `json:"hello"`
}

// handleClients returns data about known clients who have hello'd.
func handleClients(w http.ResponseWriter, r *http.Request) {
	ah := r.Header.Get("Authorization")
	if ah == "" {
		e := "client sent no API read key"
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return
	}
	af := strings.Fields(ah)
	sentKey := af[len(af)-1]
	if sentKey != readKey {
		e := "client sent wrong API read key"
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return
	}

	db, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.BusyTimeout(5 * time.Second)

	cursor, err := db.Prepare(`SELECT * FROM clients`)
	if err != nil {
		log.Println(err)
	}

	clients := make([]client, 0, 50)
	for {
		hasRow, err := cursor.Step()
		if err != nil {
			log.Println(err)
		}
		if !hasRow {
			break
		}
		var r client
		r.MachineID, _, _ = cursor.ColumnText(0)
		r.FirstSeen, _, _ = cursor.ColumnText(1)
		r.Hardware, _, _ = cursor.ColumnText(2)
		r.HostGroup, _, _ = cursor.ColumnText(3)
		r.IP, _, _ = cursor.ColumnText(4)
		r.LastSeen, _, _ = cursor.ColumnText(5)
		r.MAC, _, _ = cursor.ColumnText(6)
		r.NodeName, _, _ = cursor.ColumnText(7)
		r.OSRel, _, _ = cursor.ColumnText(8)
		r.OSSys, _, _ = cursor.ColumnText(9)
		r.OSVer, _, _ = cursor.ColumnText(10)
		r.Hello, _, _ = cursor.ColumnText(11)
		if debug {
			log.Println("sending client record", r)
		}
		clients = append(clients, r)
	}

	json.NewEncoder(w).Encode(clients)
}

// handleDoc briefly documents the API.
func handleDoc(w http.ResponseWriter, r *http.Request) {
	doc := `<!DOCTYPE html>
<html lang="en-us">
<head>
<meta charset="utf-8" />
<title>Inventory API Documentation</title>
</head>
<body>
<h1>Inventory API Documentation</h1>
<p>Clients report information about themselves (like their current IP address) through this API.</p>
<p>Administrators query the API to retrieve client information in a form useful, for example, as an Ansible inventory.</p>
<p>Clients must supply an API key in the HTTP Authorization header, and the key for reading and writing may differ. See 'example-hello.sh' and 'example-inventory.sh' supplied in the Git repository.</p>
<dl>
<dt>GET <b>/api/v1/</b> → HTML</dt>
<dd>Shows this documentation page</dd>
<dt>GET <b>/api/v1/clients</b> → JSON</dt>
<dd>Returns the full list of known clients.</dd>
<dt>PUT <b>/api/v1/hello</b> ← JSON</dt>
<dd>Client adds or updates a record about itself.</dd>
</dl>
<p>See 'README.md' for more information.</p>
<p>2019 Paul Gorman</p>
</body>
</html>`
	fmt.Fprintf(w, doc)
}

// handleHello decodes a "hello" POST from a client.
func handleHello(w http.ResponseWriter, r *http.Request) {
	ah := r.Header.Get("Authorization")
	if ah == "" {
		e := "client sent no API write key"
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return
	}
	af := strings.Fields(ah)
	sentKey := af[len(af)-1]
	if sentKey != writeKey {
		e := "client sent wrong API write key"
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return
	}

	var c client
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&c)
	if err != nil {
		log.Println(err)
	}
	if debug {
		log.Println("receiving client hello", c)
	}
	if c.MachineID == "" {
		e := "no machineID supplied"
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 400)
		return
	}

	db, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.BusyTimeout(5 * time.Second)

	err = db.Exec(`INSERT INTO clients
		(firstSeen, hardware, hostGroup, ip, lastSeen, mac, machineID, nodeName, osRel, osSys, osVer, hello)
		values (CURRENT_TIMESTAMP, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(machineID) DO UPDATE SET
		hardware=?, hostGroup=?, ip=?, lastSeen=CURRENT_TIMESTAMP, mac=?, nodeName=?, osRel=?, osSys=?, osVer=?, hello=?`,
		c.Hardware, c.HostGroup, c.IP, c.MAC, c.MachineID, c.NodeName, c.OSRel, c.OSSys, c.OSVer, c.Hello,
		c.Hardware, c.HostGroup, c.IP, c.MAC, c.NodeName, c.OSRel, c.OSSys, c.OSVer, c.Hello)
	if err != nil {
		log.Println(err)
	}
}

func init() {
	readKey = os.Getenv("readKey")
	if len(readKey) == 0 {
		log.Fatal("please set the 'readKey' environment variable")
	}
	writeKey = os.Getenv("writeKey")
	if len(writeKey) == 0 {
		log.Fatal("please set the 'writeKey' environment variable")
	}

	flag.StringVar(&addr, "address", "127.0.0.1", "network address where we server API")
	flag.IntVar(&charLimit, "char-limit", 128, "truncate JSON values supplied by clients at this limit")
	flag.StringVar(&dbFile, "db", "inventory.db", "SQLite database file")
	flag.BoolVar(&debug, "debug", false, "show verbose debugging output")
	flag.StringVar(&port, "port", "9753", "network port to serve API")
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
		err = db.Exec(`CREATE TABLE IF NOT EXISTS clients
			(machineID TEXT PRIMARY KEY, firstSeen TEXT, hardware TEXT, hostGroup TEXT,
			ip TEXT, lastSeen TEXT, mac TEXT, nodeName TEXT, osRel TEXT, osSys TEXT,
			osVer TEXT, hello TEXT)`)
		if err != nil {
			log.Println(err)
		}
	}
	if debug {
		log.Printf("listening on %s:%s", addr, port)
	}
}

func main() {
	http.HandleFunc("/", handleDoc)
	http.HandleFunc("/api/v1/clients", handleClients)
	http.HandleFunc("/api/v1/hello", handleHello)
	// "/api/v1/groups"
	// "/api/v1/inventory"
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}
