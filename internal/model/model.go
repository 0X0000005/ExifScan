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

// ScanStats 扫描统计摘要
type ScanStats struct {
	TotalCount   int            `json:"totalCount"`
	ModelDist    map[string]int `json:"modelDist"`    // 相机型号分布
	ISODist      map[string]int `json:"isoDist"`      // ISO 分布
	FNumberDist  map[string]int `json:"fNumberDist"`  // 光圈分布
	FocalLenDist map[string]int `json:"focalLenDist"` // 焦距分布
}

// ScanResult 完整扫描结果
type ScanResult struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Stats   ScanStats `json:"stats"`
	Data    []*Exif   `json:"data"`
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
