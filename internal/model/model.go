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
		Port int `yaml:"port" json:"port"`
	} `yaml:"server" json:"server"`
	Database struct {
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Driver  string `yaml:"driver" json:"driver"`
		Source  string `yaml:"source" json:"source"`
		Table   string `yaml:"table" json:"table"`
	} `yaml:"database" json:"database"`
	Scan struct {
		Path       string   `yaml:"path" json:"path"`
		Extensions []string `yaml:"extensions" json:"extensions"`
	} `yaml:"scan" json:"scan"`
	Excel struct {
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Output  string `yaml:"output" json:"output"`
	} `yaml:"excel" json:"excel"`
	Json struct {
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Output  string `yaml:"output" json:"output"`
	} `yaml:"json" json:"json"`
}
