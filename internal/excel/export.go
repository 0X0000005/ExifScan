package excel

import (
	"exifScan/internal/model"
	"fmt"

	"github.com/xuri/excelize/v2"
)

func Export(data []*model.Exif, filepath string) error {
	f := excelize.NewFile()
	sheetName := "ExifData"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// Headers
	headers := []string{"File", "Exposure Time", "ISO", "F-Number", "Focal Length", "Model", "Origin Date"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Styles
	styleOdd, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9EAF7"}, Pattern: 1},
	})
	styleEven, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFFFF"}, Pattern: 1},
	})

	// Data
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.File)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ExposureTime)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.ISO)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.FNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.FocalLength)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.Model)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.OriginDate)

		style := styleOdd
		if row%2 == 0 {
			style = styleEven
		}
		// Apply style to row
		for col := 1; col <= 7; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			f.SetCellStyle(sheetName, cell, cell, style)
		}
	}

	// Delete default Sheet1 if it exists and is not what we used
	if sheetName != "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	return f.SaveAs(filepath)
}
