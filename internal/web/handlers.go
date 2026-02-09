package web

import (
	"exifScan/internal/config"
	"exifScan/internal/db"
	"exifScan/internal/excel"
	"exifScan/internal/scan"
	"encoding/json"
	"net/http"
)

type ScanResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.AppConfig

	// 1. Scan
	results, err := scan.Scan(cfg.Scan.Path, cfg.Scan.Extensions)
	if err != nil {
		json.NewEncoder(w).Encode(ScanResponse{Success: false, Message: "Scan failed: " + err.Error()})
		return
	}

	// 2. Save to DB
	// Re-init DB in case config changed
	if err := db.InitDB(); err != nil {
		json.NewEncoder(w).Encode(ScanResponse{Success: false, Message: "DB Init failed: " + err.Error()})
		return
	}

	if err := db.Save(results); err != nil {
		json.NewEncoder(w).Encode(ScanResponse{Success: false, Message: "DB Save failed: " + err.Error()})
		return
	}

	// 3. Export to Excel
	if err := excel.Export(results, cfg.Excel.Output); err != nil {
		json.NewEncoder(w).Encode(ScanResponse{Success: false, Message: "Excel Export failed: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ScanResponse{
		Success: true,
		Message: "Scan completed successfully",
		Count:   len(results),
	})
}
