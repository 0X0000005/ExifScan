package model

type Exif struct {
	File         string `db:"file" json:"file"`
	ExposureTime string `db:"exposureTime" json:"exposureTime"`
	ISO          string `db:"iso" json:"iso"`
	FNumber      string `db:"fNumber" json:"fNumber"`
	FocalLength  string `db:"focalLength" json:"focalLength"`
	Model        string `db:"model" json:"model"`
	OriginDate   string `db:"originDate" json:"originDate"`
}

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Driver string `yaml:"driver"`
		Source string `yaml:"source"`
		Table  string `yaml:"table"`
	} `yaml:"database"`
	Scan struct {
		Path       string   `yaml:"path"`
		Extensions []string `yaml:"extensions"`
	} `yaml:"scan"`
	Excel struct {
		Output string `yaml:"output"`
	} `yaml:"excel"`
}
