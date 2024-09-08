package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

const upstreamDNS = "8.8.8.8:53" // Upstream DNS server (Google's public DNS)

// Function to handle DNS queries
func dnsQuery(w http.ResponseWriter, r *http.Request) {
	var dnsMsg []byte
	var err error

	switch r.Method {
	case "GET":
		fmt.Print("Incoming DNS request via GET")

		// Handle GET request
		queryParams := r.URL.Query()
		dnsMsgB64 := queryParams.Get("dns")
		dnsMsg, err = base64.RawURLEncoding.DecodeString(dnsMsgB64)
		if err != nil {
			http.Error(w, "Failed to decode base64 DNS query", http.StatusBadRequest)
			return
		}

	case "POST":
		fmt.Print("Incoming DNS request via POST")
		// Handle POST request
		dnsMsg, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read DNS query from body", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Only GET and POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Create a DNS message from the received query
	msg := new(dns.Msg)
	err = msg.Unpack(dnsMsg)
	if err != nil {
		http.Error(w, "Failed to unpack DNS query", http.StatusBadRequest)
		return
	}

	// Forward the DNS query to the upstream DNS server
	client := new(dns.Client)
	response, _, err := client.Exchange(msg, upstreamDNS)
	if err != nil {
		http.Error(w, "Failed to forward DNS query", http.StatusInternalServerError)
		return
	}

	// Pack the DNS response
	responseBytes, err := response.Pack()
	if err != nil {
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Encode response for GET request
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/dns-message")
		encodedResponse := base64.RawURLEncoding.EncodeToString(responseBytes)
		fmt.Fprint(w, encodedResponse)
	} else {
		// Respond with raw DNS message for POST request
		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(responseBytes)
	}
}

func main() {
	// Create HTTP server and route DNS queries to handler
	http.HandleFunc("/dns-query", dnsQuery)

	// Serve HTTPS (you'll need cert and key files for real HTTPS support)
	fmt.Println("Starting DoH server on :443")
	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
}
