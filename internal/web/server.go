package web

import (
	"embed"
	"exifScan/internal/config"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

func StartServer() {
	log.Println("Initializing Gin...")
	r := gin.Default()

	// Get sub filesystem for "static" folder to simplify paths
	fSys, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static assets
	r.StaticFS("/static", http.FS(fSys))

	// Serve HTML files directly at root/endpoints
	serve := func(c *gin.Context, file string) {
		serveEmbeddedFile(c, fSys, file)
	}

	r.GET("/", func(c *gin.Context) { serve(c, "index.html") })
	r.GET("/index.html", func(c *gin.Context) { serve(c, "index.html") })
	r.GET("/settings.html", func(c *gin.Context) { serve(c, "settings.html") })
	r.GET("/style.css", func(c *gin.Context) { serve(c, "style.css") })

	// API endpoints
	log.Println("Setting up API endpoints...")
	api := r.Group("/api")
	{
		api.GET("/config", handleGetConfig)
		api.POST("/config", handleUpdateConfig)
		api.POST("/scan", handleScan)
		api.GET("/fs/list", handleListDir)
		api.GET("/results", handleGetResults)
		api.POST("/results/import", handleImportJSON)
		api.GET("/download/excel", handleDownloadExcel)
		api.GET("/download/json", handleDownloadJSON)
	}

	port := config.AppConfig.Server.Port
	if port == 0 {
		port = 8080
	}

	log.Printf("Starting web server on port %d...", port)
	err = r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
}

func serveEmbeddedFile(c *gin.Context, fSys fs.FS, path string) {
	file, err := fSys.Open(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	http.ServeContent(c.Writer, c.Request, path, stat.ModTime(), file.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	}))
}
