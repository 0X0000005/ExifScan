package web

import (
	"exifScan/internal/config"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Initializing Gin...")
	r := gin.Default()

	// Static files
	// Static files
	staticDir := "./internal/web/static"
	if absPath, err := filepath.Abs(staticDir); err == nil {
		log.Printf("Serving static files from: %s", absPath)
	}

	r.Static("/static", staticDir) // Serve assets under /static if any

	// Serve HTML files directly at root
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "index.html"))
	})
	r.GET("/index.html", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "index.html"))
	})
	r.GET("/settings.html", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "settings.html"))
	})
	r.GET("/style.css", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "style.css"))
	})

	// API endpoints
	log.Println("Setting up API endpoints...")
	api := r.Group("/api")
	{
		api.GET("/config", handleGetConfig)
		api.POST("/config", handleUpdateConfig)
		api.POST("/scan", handleScan)
		api.GET("/fs/list", handleListDir)
	}

	port := config.AppConfig.Server.Port
	if port == 0 {
		port = 8080
	}

	log.Printf("Starting web server on port %d...", port)
	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
}
