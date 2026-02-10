package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanEmptyPath(t *testing.T) {
	_, err := Scan("", []string{".jpg"})
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestScanNonExistentPath(t *testing.T) {
	results, err := Scan("/non/existent/path/12345", []string{".jpg"})
	// 路径不存在时应返回错误或空结果
	if err == nil && len(results) > 0 {
		t.Error("expected error or empty results for non-existent path")
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exifscan_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	results, err := Scan(tmpDir, []string{".jpg", ".png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestScanSkipsNonMatchingExtensions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exifscan_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建一个 .txt 文件
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)

	results, err := Scan(tmpDir, []string{".jpg", ".png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for .txt file, got %d", len(results))
	}
}

func TestScanHandlesInvalidExif(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exifscan_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建一个假的 jpg 文件（无 EXIF 数据）
	os.WriteFile(filepath.Join(tmpDir, "fake.jpg"), []byte("not a real jpg"), 0644)

	results, err := Scan(tmpDir, []string{".jpg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该跳过无法读取 EXIF 的文件，返回空结果
	if len(results) != 0 {
		t.Errorf("expected 0 results for fake jpg, got %d", len(results))
	}
}

func TestScanSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exifscan_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建子目录
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// 在子目录中创建假图片
	os.WriteFile(filepath.Join(subDir, "test.jpg"), []byte("fake"), 0644)

	results, err := Scan(tmpDir, []string{".jpg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// fake jpg 没有 EXIF，所以结果数为 0，但不应报错
	if len(results) != 0 {
		t.Errorf("expected 0 valid EXIF results from fake jpg, got %d", len(results))
	}
}
