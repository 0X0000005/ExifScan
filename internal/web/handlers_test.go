package web

import (
	"bytes"
	"encoding/json"
	"exifScan/internal/config"
	"exifScan/internal/model"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestConfig() {
	config.AppConfig = &model.Config{}
	config.AppConfig.Server.Port = 8080
	config.AppConfig.Scan.Extensions = []string{".jpg", ".jpeg", ".png"}
	config.AppConfig.Json.Enabled = true
	config.AppConfig.Json.Output = "test_results.json"
	config.AppConfig.Excel.Enabled = false
	config.AppConfig.Database.Enabled = false
}

func setupRouter() *gin.Engine {
	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/config", handleGetConfig)
		api.POST("/config", handleUpdateConfig)
		api.POST("/scan", handleScan)
		api.GET("/fs/list", handleListDir)
		api.GET("/results", handleGetResults)
		api.POST("/results/import", handleImportJSON)
	}
	return r
}

func TestHandleGetConfig(t *testing.T) {
	setupTestConfig()
	r := setupRouter()

	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var cfg model.Config
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
}

func TestHandleUpdateConfig(t *testing.T) {
	setupTestConfig()
	r := setupRouter()

	newCfg := model.Config{}
	newCfg.Server.Port = 9090
	newCfg.Scan.Path = "/test"
	newCfg.Scan.Extensions = []string{".jpg", ".raw"}
	newCfg.Database.Driver = "mysql"
	newCfg.Database.Source = "user:pass@tcp(localhost:3306)/test"

	body, _ := json.Marshal(newCfg)
	req := httptest.NewRequest("POST", "/api/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify global config is updated
	if config.AppConfig.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.AppConfig.Server.Port)
	}
	if len(config.AppConfig.Scan.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(config.AppConfig.Scan.Extensions))
	}
	if config.AppConfig.Scan.Extensions[1] != ".raw" {
		t.Errorf("expected .raw, got %s", config.AppConfig.Scan.Extensions[1])
	}
	if config.AppConfig.Database.Driver != "mysql" {
		t.Errorf("expected mysql driver, got %s", config.AppConfig.Database.Driver)
	}
}

func TestHandleScanNoPath(t *testing.T) {
	setupTestConfig()
	config.AppConfig.Scan.Path = ""
	r := setupRouter()

	body := `{"path":""}`
	req := httptest.NewRequest("POST", "/api/scan", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleScanEmptyDir(t *testing.T) {
	setupTestConfig()
	tmpDir, err := os.MkdirTemp("", "exifscan_handler_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	defer os.Remove("test_results.json")

	r := setupRouter()

	body, _ := json.Marshal(ScanRequest{Path: tmpDir})
	req := httptest.NewRequest("POST", "/api/scan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result model.ScanResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Stats.TotalCount != 0 {
		t.Errorf("expected 0 count, got %d", result.Stats.TotalCount)
	}
}

func TestHandleListDirRoot(t *testing.T) {
	setupTestConfig()
	r := setupRouter()

	req := httptest.NewRequest("GET", "/api/fs/list?path=", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var items []FileItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	// Windows 至少有 C:
	if len(items) == 0 {
		t.Error("expected at least one drive")
	}
	// 所有条目都应该是目录
	for _, item := range items {
		if !item.IsDir {
			t.Errorf("expected only dirs, got file: %s", item.Name)
		}
	}
}

func TestHandleListDirOnlyDirs(t *testing.T) {
	setupTestConfig()
	tmpDir, err := os.MkdirTemp("", "exifscan_list_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建子目录和文件
	os.MkdirAll(tmpDir+"/subdir1", 0755)
	os.MkdirAll(tmpDir+"/subdir2", 0755)
	os.WriteFile(tmpDir+"/file.txt", []byte("hello"), 0644)
	os.WriteFile(tmpDir+"/image.jpg", []byte("fake"), 0644)

	r := setupRouter()
	req := httptest.NewRequest("GET", "/api/fs/list?path="+tmpDir, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var items []FileItem
	json.Unmarshal(w.Body.Bytes(), &items)

	if len(items) != 2 {
		t.Errorf("expected 2 dirs only, got %d items", len(items))
	}
	for _, item := range items {
		if !item.IsDir {
			t.Errorf("expected only dirs, got file: %s", item.Name)
		}
	}
}

func TestHandleImportJSON(t *testing.T) {
	setupTestConfig()
	r := setupRouter()

	// 构造 JSON 数据
	testData := []*model.Exif{
		{File: "a.jpg", Model: "Canon", ISO: "100", FNumber: "f/2.8", FocalLength: "50mm"},
		{File: "b.jpg", Model: "Sony", ISO: "400", FNumber: "f/4.0", FocalLength: "85mm"},
	}
	jsonBytes, _ := json.Marshal(testData)

	// 构造 multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(jsonBytes)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/results/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result model.ScanResult
	json.Unmarshal(w.Body.Bytes(), &result)

	if !result.Success {
		t.Error("expected success")
	}
	if result.Stats.TotalCount != 2 {
		t.Errorf("expected 2, got %d", result.Stats.TotalCount)
	}
	if result.Stats.ModelDist["Canon"] != 1 {
		t.Errorf("expected Canon=1, got %d", result.Stats.ModelDist["Canon"])
	}
}

func TestHandleGetResultsDBDisabled(t *testing.T) {
	setupTestConfig()
	config.AppConfig.Database.Enabled = false
	r := setupRouter()

	req := httptest.NewRequest("GET", "/api/results", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBuildStats(t *testing.T) {
	data := []*model.Exif{
		{Model: "Canon", ISO: "100", FNumber: "f/2.8", FocalLength: "50mm"},
		{Model: "Canon", ISO: "400", FNumber: "f/4.0", FocalLength: "85mm"},
		{Model: "Sony", ISO: "100", FNumber: "f/2.8", FocalLength: "50mm"},
		{Model: "", ISO: "", FNumber: "", FocalLength: ""},
	}

	stats := buildStats(data)

	if stats.TotalCount != 4 {
		t.Errorf("expected 4, got %d", stats.TotalCount)
	}
	if stats.ModelDist["Canon"] != 2 {
		t.Errorf("expected Canon=2, got %d", stats.ModelDist["Canon"])
	}
	if stats.ModelDist["Sony"] != 1 {
		t.Errorf("expected Sony=1, got %d", stats.ModelDist["Sony"])
	}
	if stats.ModelDist["未知"] != 1 {
		t.Errorf("expected 未知=1, got %d", stats.ModelDist["未知"])
	}
	if stats.ISODist["ISO 100"] != 2 {
		t.Errorf("expected ISO 100=2, got %d", stats.ISODist["ISO 100"])
	}
	if stats.FNumberDist["f/2.8"] != 2 {
		t.Errorf("expected f/2.8=2, got %d", stats.FNumberDist["f/2.8"])
	}
}

func TestBuildStatsEmpty(t *testing.T) {
	stats := buildStats([]*model.Exif{})
	if stats.TotalCount != 0 {
		t.Errorf("expected 0, got %d", stats.TotalCount)
	}
	if len(stats.ModelDist) != 0 {
		t.Error("expected empty ModelDist")
	}
}

func TestSaveToJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "exifscan_json_*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	data := []*model.Exif{
		{File: "test.jpg", Model: "Canon"},
	}

	if err := saveToJSON(data, tmpFile.Name()); err != nil {
		t.Fatalf("saveToJSON failed: %v", err)
	}

	// 验证文件可被读回
	f, _ := os.Open(tmpFile.Name())
	defer f.Close()
	content, _ := io.ReadAll(f)

	var loaded []*model.Exif
	if err := json.Unmarshal(content, &loaded); err != nil {
		t.Fatalf("failed to parse saved JSON: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Model != "Canon" {
		t.Error("JSON content mismatch")
	}
}
