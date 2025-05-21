package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"encoding/json"
	"time"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

)

// CORS middleware allowing all origins
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

type Config struct {
	ShouldStop bool
	MaxRuntime int // in seconds
}
func loadConfig() (*Config, error) {
	// Load .env file only in local development
	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found (probably in production)")
	}

	shouldStopStr := os.Getenv("SHOULD_STOP")
	maxRuntimeStr := os.Getenv("MAX_RUNTIME")

	fmt.Printf(".ENV  Should Stop : %s  \t Max Min : %s",shouldStopStr, maxRuntimeStr)
	// Default values
	shouldStop := true
	maxRuntime := 0

	// Parse SHOULD_STOP if set
	if shouldStopStr != "" {
		shouldStop = strings.ToLower(shouldStopStr) == "true"
	}

	// Parse MAX_RUNTIME if set and valid
	if maxRuntimeStr != "" {
		if val, err := strconv.Atoi(maxRuntimeStr); err == nil {
			maxRuntime = val
		} else {
			log.Printf("Invalid MAX_RUNTIME value: %v, using default 0\n", err)
		}
	} else {
		log.Println("MAX_RUNTIME not set, using default 0")
	}

	return &Config{
		ShouldStop: shouldStop,
		MaxRuntime: maxRuntime,
	}, nil
}

// === /should-stop ===
func shouldStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := loadConfig()
	if err != nil {
		log.Println("❌", err)
		http.Error(w, "Could not load configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%v", config.ShouldStop)
}

// === /max-runtime ===
func maxRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := loadConfig()
	if err != nil {
		log.Println("❌", err)
		http.Error(w, "Could not load configuration", http.StatusInternalServerError)
		return
	}

	log.Printf("MAX_RUNTIME: %v", config.MaxRuntime)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%d", config.MaxRuntime)
}

func updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON into map[string]interface{} to handle different types
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Load existing .env or initialize new map if file doesn't exist
	envMap, err := godotenv.Read(".env")
	if err != nil {
		envMap = make(map[string]string)
	}

	// Convert all values to string for .env compatibility
	for k, v := range updates {
		var strVal string
		switch val := v.(type) {
		case string:
			strVal = val
		case bool:
			if val {
				strVal = "true"
			} else {
				strVal = "false"
			}
		case float64:
			// JSON numbers are decoded as float64
			strVal = fmt.Sprintf("%.0f", val) // format without decimals for integers
		default:
			// Fallback: use fmt.Sprintf to convert any other type to string
			strVal = fmt.Sprintf("%v", val)
		}
		envMap[k] = strVal
	}

	// Write updated .env file
	if err := godotenv.Write(envMap, ".env"); err != nil {
		log.Println("❌ Failed to write to .env:", err)
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}

	log.Println("✅ Updated .env config successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Configuration updated successfully",
	})
}


func main() {
	http.Handle("/log", withCORS(http.HandlerFunc(logHandler)))
	http.Handle("/should-stop", withCORS(http.HandlerFunc(shouldStopHandler)))
	http.Handle("/max-runtime", withCORS(http.HandlerFunc(maxRuntimeHandler)))
	http.Handle("/update-config", withCORS(http.HandlerFunc(updateConfigHandler)))


	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
