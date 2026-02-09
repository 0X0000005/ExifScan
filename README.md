# ExifScan

ExifScan 是一个用于批量扫描图片并提取 EXIF 信息（如快门、ISO、光圈、焦段、相机型号等）的工具。
支持将数据保存到 MySQL/SQLite 数据库，并导出为 Excel 文件以便后续分析。

## 功能特性

- **批量扫描**: 支持递归扫描指定目录下的图片（支持 jpg, jpeg, arw, tiff）。
- **EXIF 提取**: 提取关键拍摄参数。
- **数据库存储**: 
  - 支持 MySQL 和 SQLite。
  - 可配置表名。
- **Excel 导出**: 自动生成带有格式的 Excel 报表。
- **Web 界面**: 提供简单的 Web 界面用于配置参数和触发扫描。

## 快速开始

### 1. 配置

修改 `config.yaml` 文件：

```yaml
server:
  port: 8080

database:
  driver: sqlite # 或 mysql
  source: exif_data.db # SQLite 文件路径 或 MySQL DSN (user:pass@tcp(host:port)/dbname)
  table: photo_data

scan:
  path: "C:/Photos" # 默认扫描路径
  extensions: [".jpg", ".jpeg", ".arw", ".tiff"]

excel:
  output: "scan_results.xlsx"
```

### 2. 运行

直接运行编译好的二进制文件：

```bash
./ExifScan.exe
# 或指定配置文件
./ExifScan.exe -config my_config.yaml
```

### 3. 使用 Web 界面

打开浏览器访问 `http://localhost:8080`。
在页面上您可以：
- 修改配置（端口、数据库、扫描路径等）。
- 点击 "Start Scan" 开始扫描。
- 查看简单的日志输出。

## 构建

需要 Go 1.20+ 环境。

```bash
# Windows
build.bat

# Linux/Mac
./build.sh
```

或者手动构建：

```bash
go mod tidy
go build -o ExifScan.exe ./cmd/exifscan
```
