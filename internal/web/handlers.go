package web

import (
	"encoding/json"
	"exifScan/internal/config"
	"exifScan/internal/db"
	"exifScan/internal/excel"
	"exifScan/internal/model"
	"exifScan/internal/scan"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// lastScanResults 保存最近一次扫描结果，用于下载
var lastScanResults []*model.Exif

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

	config.AppConfig.Database = newConfig.Database
	config.AppConfig.Scan = newConfig.Scan
	config.AppConfig.Excel = newConfig.Excel
	config.AppConfig.Json = newConfig.Json

	c.Status(http.StatusOK)
}

type ScanRequest struct {
	Path string `json:"path"`
}

func handleScan(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.Path != "" {
		config.AppConfig.Scan.Path = req.Path
	}

	cfg := config.AppConfig

	if cfg.Scan.Path == "" {
		c.JSON(http.StatusBadRequest, model.ScanResult{Success: false, Message: "扫描路径未配置"})
		return
	}

	// 1. 扫描
	results, err := scan.Scan(cfg.Scan.Path, cfg.Scan.Extensions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "扫描失败: " + err.Error()})
		return
	}

	// 保存结果供下载使用
	lastScanResults = results

	// 2. 保存到数据库（如果启用）
	if cfg.Database.Enabled {
		if err := db.InitDB(); err != nil {
			c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "数据库初始化失败: " + err.Error()})
			return
		}
		if err := db.Save(results); err != nil {
			c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "数据库保存失败: " + err.Error()})
			return
		}
	}

	// 3. 导出 Excel（如果启用）
	if cfg.Excel.Enabled {
		if err := excel.Export(results, cfg.Excel.Output); err != nil {
			c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "Excel 导出失败: " + err.Error()})
			return
		}
	}

	// 4. 导出 JSON（如果启用）
	if cfg.Json.Enabled {
		if err := saveToJSON(results, cfg.Json.Output); err != nil {
			c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "JSON 导出失败: " + err.Error()})
			return
		}
	}

	// 返回完整结果（含统计数据）
	stats := buildStats(results)
	c.JSON(http.StatusOK, model.ScanResult{
		Success: true,
		Message: "扫描完成",
		Stats:   stats,
		Data:    results,
	})
}

// handleGetResults 从数据库加载历史扫描结果
func handleGetResults(c *gin.Context) {
	if !config.AppConfig.Database.Enabled {
		c.JSON(http.StatusBadRequest, model.ScanResult{Success: false, Message: "数据库未启用"})
		return
	}

	if err := db.InitDB(); err != nil {
		c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "数据库初始化失败: " + err.Error()})
		return
	}

	results, err := db.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "查询失败: " + err.Error()})
		return
	}

	lastScanResults = results
	stats := buildStats(results)
	c.JSON(http.StatusOK, model.ScanResult{
		Success: true,
		Message: "加载历史数据成功",
		Stats:   stats,
		Data:    results,
	})
}

// handleImportJSON 导入上传的 JSON 文件
func handleImportJSON(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ScanResult{Success: false, Message: "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ScanResult{Success: false, Message: "读取文件失败: " + err.Error()})
		return
	}

	var results []*model.Exif
	if err := json.Unmarshal(data, &results); err != nil {
		c.JSON(http.StatusBadRequest, model.ScanResult{Success: false, Message: "JSON 解析失败: " + err.Error()})
		return
	}

	lastScanResults = results
	stats := buildStats(results)
	c.JSON(http.StatusOK, model.ScanResult{
		Success: true,
		Message: "导入成功",
		Stats:   stats,
		Data:    results,
	})
}

// handleDownloadExcel 下载 Excel 文件
func handleDownloadExcel(c *gin.Context) {
	output := config.AppConfig.Excel.Output
	if output == "" {
		output = "scan_results.xlsx"
	}

	// 如果有最新结果但文件不存在，重新生成
	if lastScanResults != nil {
		_ = excel.Export(lastScanResults, output)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Excel 文件不存在，请先扫描"})
		return
	}

	c.FileAttachment(output, filepath.Base(output))
}

// handleDownloadJSON 下载 JSON 文件
func handleDownloadJSON(c *gin.Context) {
	output := config.AppConfig.Json.Output
	if output == "" {
		output = "scan_results.json"
	}

	if lastScanResults != nil {
		_ = saveToJSON(lastScanResults, output)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON 文件不存在，请先扫描"})
		return
	}

	c.FileAttachment(output, filepath.Base(output))
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

// 隐藏/系统目录列表
var hiddenDirs = map[string]bool{
	"$Recycle.Bin":              true,
	"System Volume Information": true,
	"$WINDOWS.~BT":              true,
	"$WinREAgent":               true,
	"Recovery":                  true,
	"Config.Msi":                true,
}

func handleListDir(c *gin.Context) {
	pathParam := c.Query("path")

	var items []FileItem

	if pathParam == "" {
		// 列出 Windows 逻辑磁盘
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
			// 只显示目录
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			// 跳过隐藏目录（以 . 开头）和系统目录
			if strings.HasPrefix(name, ".") || hiddenDirs[name] {
				continue
			}
			items = append(items, FileItem{
				Name:  name,
				IsDir: true,
				Path:  filepath.Join(pathParam, name),
			})
		}
	}

	// 按名称排序
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	c.JSON(http.StatusOK, items)
}

// buildStats 从扫描结果构建统计数据
func buildStats(data []*model.Exif) model.ScanStats {
	stats := model.ScanStats{
		TotalCount:   len(data),
		ModelDist:    make(map[string]int),
		ISODist:      make(map[string]int),
		FNumberDist:  make(map[string]int),
		FocalLenDist: make(map[string]int),
	}

	for _, item := range data {
		if item.Model != "" {
			stats.ModelDist[item.Model]++
		} else {
			stats.ModelDist["未知"]++
		}

		if item.ISO != "" {
			stats.ISODist["ISO "+item.ISO]++
		}

		if item.FNumber != "" {
			stats.FNumberDist[item.FNumber]++
		}

		if item.FocalLength != "" {
			stats.FocalLenDist[item.FocalLength]++
		}
	}

	return stats
}
