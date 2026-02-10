package model

import (
	"testing"
)

func TestScanStatsInit(t *testing.T) {
	stats := ScanStats{
		TotalCount:   10,
		ModelDist:    map[string]int{"Canon": 5, "Sony": 3, "Nikon": 2},
		ISODist:      map[string]int{"ISO 100": 4, "ISO 400": 6},
		FNumberDist:  map[string]int{"f/2.8": 7, "f/4.0": 3},
		FocalLenDist: map[string]int{"50mm": 5, "85mm": 5},
	}

	if stats.TotalCount != 10 {
		t.Errorf("expected TotalCount=10, got %d", stats.TotalCount)
	}
	if stats.ModelDist["Canon"] != 5 {
		t.Errorf("expected Canon=5, got %d", stats.ModelDist["Canon"])
	}
	if stats.ISODist["ISO 100"] != 4 {
		t.Errorf("expected ISO 100=4, got %d", stats.ISODist["ISO 100"])
	}
}

func TestScanResultInit(t *testing.T) {
	result := ScanResult{
		Success: true,
		Message: "ok",
		Stats:   ScanStats{TotalCount: 3},
		Data: []*Exif{
			{File: "a.jpg", Model: "Canon"},
			{File: "b.jpg", Model: "Sony"},
			{File: "c.jpg", Model: "Canon"},
		},
	}

	if !result.Success {
		t.Error("expected success=true")
	}
	if len(result.Data) != 3 {
		t.Errorf("expected 3 data items, got %d", len(result.Data))
	}
	if result.Stats.TotalCount != 3 {
		t.Errorf("expected TotalCount=3, got %d", result.Stats.TotalCount)
	}
}

func TestExifStruct(t *testing.T) {
	e := &Exif{
		File:         "/photos/test.jpg",
		ExposureTime: "1/125",
		ISO:          "200",
		FNumber:      "f/2.8",
		FocalLength:  "50mm",
		Model:        "Canon EOS R5",
		OriginDate:   "2024-01-01 12:00:00",
	}

	if e.File != "/photos/test.jpg" {
		t.Errorf("unexpected File: %s", e.File)
	}
	if e.Model != "Canon EOS R5" {
		t.Errorf("unexpected Model: %s", e.Model)
	}
}
