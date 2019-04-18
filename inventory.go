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
var requestByteLimit int64
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

// handle404 returns documentation about the API.
func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	doc := `<!DOCTYPE html><html lang="en-us"><head><meta charset="utf-8" /><pre>

Inventory API Documentation

Clients report information about themselves (like their current IP address)
through this API.

Administrators query the API to retrieve client information in a form useful,
for example, as an Ansible inventory.

Clients must supply an API key in the HTTP Authorization header, and the key
for reading and writing may differ. See 'example-hello.sh' and
'example-inventory.sh' supplied in the Git repository.

GET  /api/v1/  → HTML
	Shows this documentation page

GET  /api/v1/clients  → JSON
	Returns the full list of known clients.

PUT  /api/v1/hello  ← JSON
	Client adds or updates a record about itself.

GET  /api/v1/inventory  → INI
	Returns a list of known clients in INI format, usable as an Ansible
	inventory file.

See 'README.md' for more information.

2019 Paul Gorman

</pre></body></html>`
	fmt.Fprintf(w, doc)
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

	conn, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	conn.BusyTimeout(5 * time.Second)

	stmt, err := conn.Prepare(`SELECT * FROM clients`)
	if err != nil {
		log.Println(err)
	}

	clients := make([]client, 0, 50)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Println(err)
		}
		if !hasRow {
			break
		}
		clients = append(clients, unpackClient(stmt))
	}

	if debug {
		log.Println("sending client records", clients)
	}
	json.NewEncoder(w).Encode(clients)
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
	br := http.MaxBytesReader(w, r.Body, requestByteLimit)
	decoder := json.NewDecoder(br)
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

// handleInventory returns data about known clients in a format suitable for Ansible.
func handleInventory(w http.ResponseWriter, r *http.Request) {
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

	conn, err := sqlite3.Open(dbFile)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()
	conn.BusyTimeout(5 * time.Second)

	stmt, err := conn.Prepare(`SELECT * FROM clients`)
	if err != nil {
		log.Println(err)
	}

	clients := make([]client, 0, 50)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Println(err)
		}
		if !hasRow {
			break
		}
		clients = append(clients, unpackClient(stmt))
	}

	// TODO Print hosts by group

	if debug {
		log.Println("sending client records", clients)
	}
	for _, c := range clients {
		fmt.Fprintln(w, c.IP)
	}
}

// unpackClient fill a client struct with data from a SQLite row.
func unpackClient(stmt *sqlite3.Stmt) client {
	var c client
	c.MachineID, _, _ = stmt.ColumnText(0)
	c.FirstSeen, _, _ = stmt.ColumnText(1)
	c.Hardware, _, _ = stmt.ColumnText(2)
	c.HostGroup, _, _ = stmt.ColumnText(3)
	c.IP, _, _ = stmt.ColumnText(4)
	c.LastSeen, _, _ = stmt.ColumnText(5)
	c.MAC, _, _ = stmt.ColumnText(6)
	c.NodeName, _, _ = stmt.ColumnText(7)
	c.OSRel, _, _ = stmt.ColumnText(8)
	c.OSSys, _, _ = stmt.ColumnText(9)
	c.OSVer, _, _ = stmt.ColumnText(10)
	c.Hello, _, _ = stmt.ColumnText(11)
	return c
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
	flag.Int64Var(&requestByteLimit, "byte-limit", 512, "limit bytes sent by clients to this maximium")
	flag.StringVar(&dbFile, "db", "inventory.db", "SQLite database file")
	flag.BoolVar(&debug, "debug", false, "show verbose debugging output")
	flag.StringVar(&port, "port", "9753", "network port to serve API")
	flag.Parse()

	_, err := os.Stat(dbFile)
	if err != nil {
		log.Println("database file", dbFile, "doesn't already exist")
		conn, err := sqlite3.Open(dbFile)
		if err != nil {
			log.Println(err)
		}
		defer conn.Close()
		conn.BusyTimeout(5 * time.Second)
		err = conn.Exec(`CREATE TABLE IF NOT EXISTS clients
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
	http.HandleFunc("/", handle404)
	http.HandleFunc("/api/v1/clients", handleClients)
	http.HandleFunc("/api/v1/hello", handleHello)
	http.HandleFunc("/api/v1/inventory", handleInventory)
	// "/api/v1/groups"
	s := &http.Server{
		Addr:           addr + ":" + port,
		MaxHeaderBytes: 1 << 4,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
