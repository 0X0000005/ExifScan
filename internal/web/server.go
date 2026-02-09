package web

import (
	"encoding/json"
	"exifScan/internal/config"
	"exifScan/internal/model"
	"fmt"
	"log"
	"net/http"
)

func StartServer() {
	mux := http.NewServeMux()

	// Static files
	// In production, you might want to embed these using `embed` package
	// For now, serving from disk for simplicity during dev, or embed if requested.
	// Plan said "Frontend will be embedded in binary". So let's prepare for embed,
	// but for now I'll serve from directory relative to binary or just code it.
	// Actually, I'll just serve the static directory.

	fs := http.FileServer(http.Dir("internal/web/static"))
	mux.Handle("/", fs)

	// API endpoints
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/scan", handleScan)

	port := config.AppConfig.Server.Port
	if port == 0 {
		port = 8080
	}

	log.Printf("Starting web server on port %d...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatal(err)
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(config.AppConfig)
	} else if r.Method == http.MethodPost {
		// Update config logic (simplified, in-memory updates for run)
		// For persistence, we'd need to write back to config.yaml
		// Here we just update the in-memory struct
		var newConfig model.Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Carefully update allowlisted fields
		config.AppConfig.Database = newConfig.Database
		config.AppConfig.Scan = newConfig.Scan
		config.AppConfig.Excel = newConfig.Excel

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
