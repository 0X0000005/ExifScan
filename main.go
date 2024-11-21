package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "scan path")
	mode := flag.String("mode", "", "scan mode")
	userName := flag.String("u", "", "username")
	password := flag.String("p", "", "password")
	url := flag.String("url", "", "url")
	flag.Parse()
	var record func(exifInfo *Exif) error
	if *mode == "database" {
		defer func() {
			if db != nil {
				db.Close()
			}
		}()
		err := connectMysql(*userName, *password, *url)
		if err != nil {
			panic(err)
		}
		record = insertMysql
	} else if *mode == "excel" {
		initExcel()
		record = insertExcel
	} else {
		panic(errors.New("unknown mode"))
	}
	err := scan(*dir, record)
	if err != nil {
		panic(err)
	}

}

var db *sqlx.DB

func connectMysql(userName, password, url string) error {
	if userName == "" || password == "" || url == "" {
		panic("username or password or url is empty")
	}
	dsn := fmt.Sprintf("%s:%s@%s", userName, password, url)
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	return nil
}

var file *excelize.File

var sheetName string

var styleOdd int

var styleEven int

func initExcel() {
	file = excelize.NewFile()
	// 设置工作表名称
	sheetName = "ExifData"
	file.NewSheet(sheetName)
	// 设置表头
	file.SetCellValue(sheetName, "A1", "File")
	file.SetCellValue(sheetName, "B1", "Exposure Time")
	file.SetCellValue(sheetName, "C1", "ISO")
	file.SetCellValue(sheetName, "D1", "F-Number")
	file.SetCellValue(sheetName, "E1", "Focal Length")
	file.SetCellValue(sheetName, "F1", "Model")
	file.SetCellValue(sheetName, "G1", "Origin Date")

	styleOdd, _ = file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D9EAF7"},
			Pattern: 1,
		},
	})
	styleEven, _ = file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
	})
	save()
}

func save() error {
	if err := file.SaveAs("ExifDataWithColors.xlsx"); err != nil {
		fmt.Println("Error saving file:", err)
		return err
	}
	return nil
}

func getNextRow() (int, error) {
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 2, nil
	}
	return len(rows) + 1, nil
}

func insertMysql(e *Exif) error {
	query := `
		INSERT INTO photo (file, exposureTime, iso, fNumber, focalLength, model, originDate)
		VALUES (:file, :exposureTime, :iso, :fNumber, :focalLength, :model, :originDate)
	`
	_, err := db.NamedExec(query, e)
	return err
}

func insertExcel(e *Exif) error {
	columns := []string{"A", "B", "C", "D", "E", "F", "G"}
	row, err := getNextRow()
	if err != nil {
		return err
	}
	// 插入 Exif 数据到对应的单元格
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[0], row), e.File)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[1], row), e.ExposureTime)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[2], row), e.ISO)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[3], row), e.FNumber)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[4], row), e.FocalLength)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[5], row), e.Model)
	if err != nil {
		return err
	}
	err = file.SetCellValue(sheetName, fmt.Sprintf("%s%d", columns[6], row), e.OriginDate)
	if err != nil {
		return err
	}
	var style int
	if row%2 == 0 {
		style = styleEven
	} else {
		style = styleOdd
	}

	// 应用样式到当前行的所有单元格
	for _, col := range columns {
		cell := fmt.Sprintf("%s%d", col, row)
		if err := file.SetCellStyle(sheetName, cell, cell, style); err != nil {
			return err
		}
	}
	return save()
}

func scan(path string, record func(*Exif) error) error {
	if path == "" {
		return errors.New("path is empty")
	}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.ToLower(filepath.Ext(path)) != ".arw" {
			return nil
		}
		if info.IsDir() {
			// 如果是目录，递归调用 scan 并获取该目录下的 EXIF 数据
			return nil
		} else {
			// 如果是文件，调用 getExif 获取 EXIF 数据
			log.Printf("获取文件[%s]的 EXIF 数据...", path)
			ex, err := getExif(path)
			if err != nil {
				return err
			}
			err = record(ex)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func getExif(path string) (*Exif, error) {
	ex := &Exif{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	e, err := exif.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	ex.File = path
	// 获取曝光时间（ExposureTime）
	if exposure, err := e.Get(exif.ExposureTime); err == nil {
		exp := exposure.String()
		exp = strings.Replace(exp, "\"", "", -1)
		ex.ExposureTime = exp
	} else {
		return nil, err
	}

	// 获取 ISO 值
	if iso, err := e.Get(exif.ISOSpeedRatings); err == nil {
		ex.ISO = iso.String()
	} else {
		return nil, err
	}

	// 获取并格式化光圈值（FNumber）
	if fNumber, err := e.Get(exif.FNumber); err == nil {
		// 将光圈值格式化为 "f/x" 格式
		if f, err := fNumber.Rat(0); err == nil {
			num, _ := f.Float64()
			ex.FNumber = fmt.Sprintf("f/%.1f", num)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 获取并格式化焦距值（FocalLength）
	if focalLength, err := e.Get(exif.FocalLength); err == nil {
		// 将焦距值格式化为 "XXmm" 格式
		if f, err := focalLength.Rat(0); err == nil {
			num, b := f.Float64()
			if !b {
				return nil, errors.New("focal Length Error")
			}
			ex.FocalLength = fmt.Sprintf("%.0fmm", num)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	if model, err := e.Get(exif.Model); err == nil {
		m, err := model.StringVal()
		if err != nil {
			return nil, err
		}
		ex.Model = m
	} else {
		return nil, err
	}

	if time, err := e.DateTime(); err == nil {
		ex.OriginDate = time.Format("2006-01-02 15:04:05")
	} else {
		return nil, err
	}
	return ex, nil
}

type Exif struct {
	File         string `db:"file"`
	ExposureTime string `db:"exposureTime"`
	ISO          string `db:"iso"`
	FNumber      string `db:"fNumber"`
	FocalLength  string `db:"focalLength"`
	Model        string `db:"model"`
	OriginDate   string `db:"originDate"`
}
