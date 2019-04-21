package main

import (
	"log"
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		header, value, want string
	}{
		// https://tools.ietf.org/html/rfc7239
		{"Forwarded", `for="_gazonk"`, "_gazonk"},
		{"Forwarded", `For="[2001:db8:cafe::17]:4711"`, "[2001:db8:cafe::17]:4711"},
		{"Forwarded", `for=192.0.2.60;proto=http;by=203.0.113.43`, "192.0.2.60"},
		{"Forwarded", `for=192.0.2.43, for=198.51.100.17`, "192.0.2.43"},
		{"Forwarded", `😈`, ""},
		{"X-Real-IP", "192.0.2.99", "192.0.2.99"},
		{"X-Real-IP", `😈`, "😈"},
		{"X-Forwarded-For", "192.0.2.29, 10.0.0.200, 192.168.3.4", "192.0.2.29"},
		{"X-Forwarded-For", "192.0.2.19,10.0.0.200,192.168.3.4", "192.0.2.19"},
		{"", "", "192.0.2.1"},
	}
	for _, c := range cases {
		r, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			log.Println(err)
		}
		r.RemoteAddr = "192.0.2.1"
		r.Header.Add(c.header, c.value)

		got := clientIP(r)
		if got != c.want {
			t.Errorf("for header '%s: %s' clientIP → %q, but we want %q", c.header, c.value, got, c.want)
		}
	}
}
