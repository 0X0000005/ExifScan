package db

import (
	"exifScan/internal/config"
	"exifScan/internal/model"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB() error {
	cfg := config.AppConfig.Database
	var dsn string
	if cfg.Driver == "mysql" {
		dsn = cfg.Source
	} else if cfg.Driver == "sqlite" || cfg.Driver == "sqlite3" {
		dsn = cfg.Source
		cfg.Driver = "sqlite3" // Ensure correct driver name for sqlx
	} else {
		return fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	var err error
	DB, err = sqlx.Connect(cfg.Driver, dsn)
	if err != nil {
		return err
	}

	// Create table if not exists
	var schema string
	if cfg.Driver == "mysql" {
		schema = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				file VARCHAR(255),
				exposureTime VARCHAR(50),
				iso VARCHAR(50),
				fNumber VARCHAR(50),
				focalLength VARCHAR(50),
				model VARCHAR(100),
				originDate VARCHAR(50)
			);
		`, cfg.Table)
	} else {
		// SQLite syntax
		schema = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				file TEXT,
				exposureTime TEXT,
				iso TEXT,
				fNumber TEXT,
				focalLength TEXT,
				model TEXT,
				originDate TEXT
			);
		`, cfg.Table)
	}

	_, err = DB.Exec(schema)
	return err
}

func Save(data []*model.Exif) error {
	if len(data) == 0 {
		return nil
	}
	cfg := config.AppConfig.Database
	query := fmt.Sprintf(`
		INSERT INTO %s (file, exposureTime, iso, fNumber, focalLength, model, originDate)
		VALUES (:file, :exposureTime, :iso, :fNumber, :focalLength, :model, :originDate)
	`, cfg.Table)

	// Batch insert could be better, but NamedExec is easier for now
	// To do batch with sqlx, usually loop or struct scan.
	// For simplicity and matching previous logic style:
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	for _, item := range data {
		_, err := tx.NamedExec(query, item)
		if err != nil {
			log.Printf("Error inserting %s: %v", item.File, err)
			// Decide if we abort or continue. Let's continue but log.
			continue
		}
	}

	return tx.Commit()
}

// Query 从数据库加载所有扫描记录
func Query() ([]*model.Exif, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	table := config.AppConfig.Database.Table
	var results []*model.Exif
	err := DB.Select(&results, fmt.Sprintf("SELECT file, exposureTime, iso, fNumber, focalLength, model, originDate FROM %s", table))
	return results, err
}
