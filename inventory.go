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
	"regexp"
	"strings"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var addr string
var dbFile string
var debug bool
var readKey string
var reForwarded *regexp.Regexp
var requestByteLimit int64
var port string
var writeKey string

type client struct {
	FirstSeen string `json:"firstSeen"`
	Hello     string `json:"hello"`
	HostGroup string `json:"hostGroup"`
	IP        string `json:"ip"`
	LastSeen  string `json:"lastSeen"`
	MachineID string `json:"machineID"`
	NodeName  string `json:"nodeName"`
	Variables string `json:"variables"`
}

// checkAPIKey check the API key sent by the client.
func checkAPIKey(w http.ResponseWriter, r *http.Request) bool {
	var k string
	if r.Method == "GET" {
		k = readKey
	} else {
		k = writeKey
	}
	ah := r.Header.Get("Authorization")
	if ah == "" {
		e := fmt.Sprintf("%s sent server no API key for %s %v", clientIP(r), r.Method, r.URL)
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return false
	}
	af := strings.Fields(ah)
	sentKey := af[len(af)-1]
	if sentKey != k {
		e := fmt.Sprintf("%s sent server the wrong API key for %s %v", clientIP(r), r.Method, r.URL)
		if debug {
			log.Println(e)
		}
		http.Error(w, e, 401)
		return false
	}
	return true
}

// clientIP sets the IP address written to logs and debugging messages.
func clientIP(r *http.Request) string {
	var addr string
	if r.Header.Get("Forwarded") != "" { // RFC 7239
		if reForwarded.FindStringSubmatch(r.Header.Get("Forwarded")) != nil {
			addr = reForwarded.FindStringSubmatch(r.Header.Get("Forwarded"))[1]
		}
	} else if r.Header.Get("X-Real-IP") != "" {
		addr = r.Header.Get("X-Real-IP")
	} else if r.Header.Get("X-Forwarded-For") != "" {
		addr = strings.SplitN(r.Header.Get("X-Forwarded-For"), ",", 2)[0]
	} else {
		addr = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}
	return addr
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
	if !checkAPIKey(w, r) {
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
		log.Println(clientIP(r), r.Method, r.RequestURI)
	}
	json.NewEncoder(w).Encode(clients)
}

// handleHello decodes a "hello" POST from a client.
func handleHello(w http.ResponseWriter, r *http.Request) {
	if !checkAPIKey(w, r) {
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
		log.Println(clientIP(r), r.Method, r.RequestURI, c)
	}
	if c.MachineID == "" {
		e := fmt.Sprintf("%s supplied no machineID %s %v", clientIP(r), r.Method, r.RequestURI)
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
		(firstSeen, hello, hostGroup, ip, lastSeen, machineID, nodeName, variables)
		values (CURRENT_TIMESTAMP, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?)
		ON CONFLICT(machineID) DO UPDATE SET
		hello=?, hostGroup=?, ip=?, lastSeen=CURRENT_TIMESTAMP, nodeName=?, variables=?`,
		c.Hello, c.HostGroup, c.IP, c.MachineID, c.NodeName, c.Variables,
		c.Hello, c.HostGroup, c.IP, c.NodeName, c.Variables)
	if err != nil {
		log.Println(err)
	}
}

// handleInventory returns data about known clients in a format suitable for Ansible.
func handleInventory(w http.ResponseWriter, r *http.Request) {
	if !checkAPIKey(w, r) {
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

	groups := make([]string, 0, 20)
	for _, c := range clients {
		found := false
		for _, g := range groups {
			if g == c.HostGroup {
				found = true
				break
			}
		}
		if found == false {
			groups = append(groups, c.HostGroup)
		}
	}

	if debug {
		log.Println(clientIP(r), r.Method, r.RequestURI)
	}
	for _, c := range clients {
		if c.HostGroup == "" {
			fmt.Fprintln(w, c.IP, c.Variables)
		}
	}
	for _, g := range groups {
		if g == "" {
			continue
		}
		fmt.Fprintf(w, "\n[%s]\n", g)
		for _, c := range clients {
			if c.HostGroup == g {
				fmt.Fprintln(w, c.IP, c.Variables)
			}
		}
	}
}

// unpackClient fill a client struct with data from a SQLite row.
func unpackClient(stmt *sqlite3.Stmt) client {
	var c client
	c.FirstSeen, _, _ = stmt.ColumnText(0)
	c.Hello, _, _ = stmt.ColumnText(1)
	c.HostGroup, _, _ = stmt.ColumnText(2)
	c.IP, _, _ = stmt.ColumnText(3)
	c.LastSeen, _, _ = stmt.ColumnText(4)
	c.MachineID, _, _ = stmt.ColumnText(5)
	c.NodeName, _, _ = stmt.ColumnText(6)
	c.Variables, _, _ = stmt.ColumnText(7)
	return c
}

func init() {
	reForwarded = regexp.MustCompile(`(?i)for="?([^,;\s"]+)"?`)

	readKey = os.Getenv("readKey")
	if len(readKey) == 0 {
		log.Fatal("please set the 'readKey' environment variable")
	}
	writeKey = os.Getenv("writeKey")
	if len(writeKey) == 0 {
		log.Fatal("please set the 'writeKey' environment variable")
	}

	flag.StringVar(&addr, "address", "127.0.0.1", "network address where we server API")
	flag.Int64Var(&requestByteLimit, "byte-limit", 16000, "limit bytes sent by clients to this maximium")
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
			(firstSeen TEXT, hello TEXT, hostGroup TEXT, ip TEXT, lastSeen TEXT,
			machineID TEXT PRIMARY KEY, nodeName TEXT, variables TEXT)`)
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
