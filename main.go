// sub2api - A subscription converter API service
// Fork of Wei-Shaw/sub2api
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yourusername/sub2api/handler"
)

const (
	defaultHost    = "127.0.0.1" // changed from 0.0.0.0 - prefer localhost-only by default for personal use
	defaultPort    = 8080
	appName        = "sub2api"
	appVersion     = "dev"
)

func main() {
	var (
		host    string
		port    int
		version bool
	)

	flag.StringVar(&host, "host", getEnvOrDefault("HOST", defaultHost), "Host address to listen on")
	flag.IntVar(&port, "port", getEnvIntOrDefault("PORT", defaultPort), "Port to listen on")
	flag.BoolVar(&version, "version", false, "Print version information and exit")
	flag.Parse()

	if version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	router := handler.NewRouter()

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Starting %s on %s", appName, addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second, // reduced from 30s - 15s is plenty for local use
		WriteTimeout: 60 * time.Second, // increased from 30s - some subscriptions can be slow to fetch
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultValue
}

// getEnvIntOrDefault returns the integer value of an environment variable or a default value.
func getEnvIntOrDefault(key string, defaultValue int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		log.Printf("Warning: invalid integer value for %s, using default %d", key, defaultValue)
	}
	return defaultValue
}
