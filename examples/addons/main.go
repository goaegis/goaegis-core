package main

import (
	"log"
	"time"

	aegis "github.com/goaegis/goaegis-core/aegis/core"
	"github.com/goaegis/goaegis-core/examples/addons/logging"
	"github.com/goaegis/goaegis-core/examples/addons/remote"
)

func main() {
	log.Println("=== goaegis Addon Examples ===")

	// Create new Aegis instance
	authz := aegis.New()
	defer authz.Shutdown()

	// Register logging addon (verbose mode)
	log.Println("\n1. Registering logging addon...")
	loggingAddon := logging.New(true)
	if err := authz.Use(loggingAddon); err != nil {
		log.Fatal(err)
	}

	// Register S3 config loader addon (simulated with local file)
	log.Println("\n2. Registering S3 config loader addon...")
	s3Addon := remote.New("my-bucket", "config/auth.yaml", 5*time.Second, "../simple/config.yaml")
	if err := authz.Use(s3Addon); err != nil {
		log.Fatal(err)
	}

	// Load config (S3 addon will provide the source)
	log.Println("\n3. Loading config from S3 (simulated)...")
	if err := authz.LoadConfigFromAddon(); err != nil {
		log.Fatal(err)
	}

	// Test authorization
	log.Println("\n4. Testing authorization...")
	testCases := []struct {
		subject  string
		resource string
		action   string
	}{
		{"user:alice", "posts", "read"},
		{"user:bob", "posts", "create"},
		{"user:alice", "posts", "delete"},
	}

	for _, tc := range testCases {
		allowed, err := authz.Can(tc.subject, tc.resource, tc.action, nil)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		result := "DENIED"
		if allowed {
			result = "ALLOWED"
		}
		log.Printf("%s: %s -> %s.%s", result, tc.subject, tc.resource, tc.action)
	}

	// Simulate hot reload
	log.Println("\n5. Waiting for config changes (modify ../simple/config.yaml)...")
	log.Println("   Press Ctrl+C to exit")

	// Keep running to demonstrate hot reload
	select {}
}
