package scan

import (
	"bytes"
	"exifScan/internal/model"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

func Scan(path string, extensions []string) ([]*model.Exif, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}
	var results []*model.Exif
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		matched := false
		for _, e := range extensions {
			if strings.ToLower(e) == ext {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		log.Printf("Scanning file: %s", path)
		ex, err := getExif(path)
		if err != nil {
			log.Printf("Failed to get EXIF for %s: %v", path, err)
			// Continue scanning other files even if one fails
			return nil
		}
		results = append(results, ex)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func getExif(path string) (*model.Exif, error) {
	ex := &model.Exif{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	e, err := exif.Decode(bytes.NewReader(b))
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
