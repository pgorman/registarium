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
	flag.StringVar(&dbFile, "db", "simpleinventory.db", "SQLite database file")
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
	http.HandleFunc("/api/v1/clients", handleClients)
	http.HandleFunc("/api/v1/hello", handleHello)
	// "/api/v1/groups"
	// "/api/v1/inventory"
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}
