package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)
// Shared variable to control stop state
var shouldStop = true

// CORS middleware allowing all origins
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Open CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// POST handler to log content with timestamp
func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	content := string(body)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	log.Printf("[%s] Received: %s\n", timestamp, content)

	fmt.Fprintf(w, "Received")
}

// GET handler to return the current stop state
func shouldStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v", shouldStop)
}

func main() {
	http.Handle("/log", withCORS(http.HandlerFunc(logHandler)))
	http.Handle("/should-stop", withCORS(http.HandlerFunc(shouldStopHandler)))

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
