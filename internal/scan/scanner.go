package scan

import (
	"errors"
	"exifScan/internal/model"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/rwcarlsen/goexif/exif"
)

// maxExifRead 只读取文件前 64KB，EXIF 头部通常在文件开头
const maxExifRead = 64 * 1024

func Scan(path string, extensions []string) ([]*model.Exif, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	// 收集所有匹配的文件路径
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// 跳过无法访问的目录，继续扫描
			if info != nil && info.IsDir() {
				log.Printf("跳过无法访问的目录: %s", p)
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		for _, e := range extensions {
			if strings.ToLower(e) == ext {
				files = append(files, p)
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []*model.Exif{}, nil
	}

	// 并发扫描
	workers := runtime.NumCPU()
	if workers > 8 {
		workers = 8
	}
	if workers < 2 {
		workers = 2
	}

	type result struct {
		exif *model.Exif
		err  error
	}

	fileCh := make(chan string, len(files))
	resultCh := make(chan result, len(files))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileCh {
				ex, err := getExif(f)
				resultCh <- result{exif: ex, err: err}
			}
		}()
	}

	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []*model.Exif
	for r := range resultCh {
		if r.err != nil {
			log.Printf("读取 EXIF 失败: %v", r.err)
			continue
		}
		results = append(results, r.exif)
	}

	log.Printf("扫描完成，共发现 %d 个文件，成功提取 %d 个 EXIF", len(files), len(results))
	return results, nil
}

func getExif(path string) (*model.Exif, error) {
	ex := &model.Exif{}

	// 只读取文件前 64KB，EXIF 数据通常在文件头部
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := io.LimitReader(f, maxExifRead)
	e, err := exif.Decode(reader)
	if err != nil {
		return nil, err
	}
	ex.File = path

	// ExposureTime
	if exposure, err := e.Get(exif.ExposureTime); err == nil {
		exp := exposure.String()
		exp = strings.Replace(exp, "\"", "", -1)
		ex.ExposureTime = exp
	}

	// ISO
	if iso, err := e.Get(exif.ISOSpeedRatings); err == nil {
		ex.ISO = iso.String()
	}

	// FNumber
	if fNumber, err := e.Get(exif.FNumber); err == nil {
		if f, err := fNumber.Rat(0); err == nil {
			num, _ := f.Float64()
			ex.FNumber = fmt.Sprintf("f/%.1f", num)
		}
	}

	// FocalLength
	if focalLength, err := e.Get(exif.FocalLength); err == nil {
		if f, err := focalLength.Rat(0); err == nil {
			num, b := f.Float64()
			if b {
				ex.FocalLength = fmt.Sprintf("%.0fmm", num)
			}
		}
	}

	// Model
	if modelVal, err := e.Get(exif.Model); err == nil {
		m, err := modelVal.StringVal()
		if err == nil {
			ex.Model = m
		}
	}

	// OriginDate
	if time, err := e.DateTime(); err == nil {
		ex.OriginDate = time.Format("2006-01-02 15:04:05")
	}

	return ex, nil
}
