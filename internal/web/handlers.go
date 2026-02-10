package web

import (
	"encoding/json"
	"exifScan/internal/config"
	"exifScan/internal/db"
	"exifScan/internal/excel"
	"exifScan/internal/model"
	"exifScan/internal/scan"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type ScanResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

type FileItem struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Path  string `json:"path"`
}

func handleGetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, config.AppConfig)
}

func handleUpdateConfig(c *gin.Context) {
	var newConfig model.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update in-memory config
	config.AppConfig.Database = newConfig.Database
	config.AppConfig.Scan = newConfig.Scan
	config.AppConfig.Excel = newConfig.Excel
	config.AppConfig.Json = newConfig.Json

	// TODO: Persist to file if needed for next restart

	c.Status(http.StatusOK)
}

type ScanRequest struct {
	Path string `json:"path"`
}

func handleScan(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.Path != "" {
		// Use path from request
		config.AppConfig.Scan.Path = req.Path
	}

	cfg := config.AppConfig

	if cfg.Scan.Path == "" {
		c.JSON(http.StatusBadRequest, ScanResponse{Success: false, Message: "Scan path not configured"})
		return
	}

	// 1. Scan
	results, err := scan.Scan(cfg.Scan.Path, cfg.Scan.Extensions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ScanResponse{Success: false, Message: "Scan failed: " + err.Error()})
		return
	}

	// 2. Save to DB (if enabled)
	if cfg.Database.Enabled {
		if err := db.InitDB(); err != nil {
			c.JSON(http.StatusInternalServerError, ScanResponse{Success: false, Message: "DB Init failed: " + err.Error()})
			return
		}

		if err := db.Save(results); err != nil {
			c.JSON(http.StatusInternalServerError, ScanResponse{Success: false, Message: "DB Save failed: " + err.Error()})
			return
		}
	}

	// 3. Export to Excel (if enabled)
	if cfg.Excel.Enabled {
		if err := excel.Export(results, cfg.Excel.Output); err != nil {
			c.JSON(http.StatusInternalServerError, ScanResponse{Success: false, Message: "Excel Export failed: " + err.Error()})
			return
		}
	}

	// 4. Export to JSON (if enabled)
	if cfg.Json.Enabled {
		if err := saveToJSON(results, cfg.Json.Output); err != nil {
			c.JSON(http.StatusInternalServerError, ScanResponse{Success: false, Message: "JSON Export failed: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, ScanResponse{
		Success: true,
		Message: "Scan completed successfully",
		Count:   len(results),
	})
}

func saveToJSON(data []*model.Exif, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func handleListDir(c *gin.Context) {
	pathParam := c.Query("path")

	var items []FileItem

	if pathParam == "" {
		// List logic drives on Windows
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			drivePath := string(drive) + ":\\"
			if _, err := os.Stat(drivePath); err == nil {
				items = append(items, FileItem{Name: string(drive) + ":", IsDir: true, Path: drivePath})
			}
		}
	} else {
		entries, err := os.ReadDir(pathParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, entry := range entries {
			// Skip hidden files/dirs if needed, for now include all
			items = append(items, FileItem{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
				Path:  filepath.Join(pathParam, entry.Name()),
			})
		}
	}

	// Sort directories first
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir && !items[j].IsDir {
			return true
		}
		if !items[i].IsDir && items[j].IsDir {
			return false
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	c.JSON(http.StatusOK, items)
}
