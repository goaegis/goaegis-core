package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	aegis "github.com/dovakiin0/goaegis-core/aegis/core"
	"github.com/dovakiin0/goaegis-core/aegis/middleware"
)

var authz *aegis.Aegis

func main() {
	// Get config path from environment or use default
	configPath := os.Getenv("AEGIS_CONFIG_PATH")
	if configPath == "" {
		configPath = "./config"
	}

	// Initialize goaegis
	authz = aegis.New()
	if err := authz.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("✅ goaegis configuration loaded successfully")

	// Setup routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/authorize", authorizeHandler)

	// Protected route example
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/admin/settings", adminSettingsHandler)

	// Wrap with authorization middleware
	// Extract subject from header (in production, extract from JWT/session)
	subjectExtractor := func(r *http.Request) string {
		return r.Header.Get("X-Subject-ID")
	}

	http.Handle("/admin/settings",
		middleware.Require(authz, subjectExtractor, "settings", "update")(protectedMux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 goaegis-server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "goaegis-server",
	})
}

// authorizeHandler provides a REST endpoint for authorization checks
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Subject  string                 `json:"subject"`
		Resource string                 `json:"resource"`
		Action   string                 `json:"action"`
		Context  map[string]interface{} `json:"context,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	allowed, err := authz.Can(req.Subject, req.Resource, req.Action, req.Context)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authorization error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"allowed":  allowed,
		"subject":  req.Subject,
		"resource": req.Resource,
		"action":   req.Action,
	})
}

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to admin settings - you are authorized!")
}
