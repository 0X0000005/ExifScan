# ExifScan

ExifScan 是一个用于批量扫描图片并提取 EXIF 信息的工具，支持 **可视化图表分析**。

提取快门、ISO、光圈、焦段、相机型号等关键拍摄参数，生成直观的统计图表，帮助摄影师了解自己的拍摄习惯。

## ✨ 功能特性

### 扫描与分析
- **批量扫描**: 递归扫描目录下的图片（支持 `.jpg` `.jpeg` `.png` `.arw` `.tiff`）
- **并发处理**: 多核并行 EXIF 提取，大目录扫描速度快
- **EXIF 提取**: 提取快门速度、ISO、光圈、焦距、相机型号、拍摄日期

### 📊 图表可视化
- **相机型号分布**: 饼图展示各型号使用占比
- **ISO 分布**: 柱状图展示常用 ISO 段
- **光圈分布**: 柱状图展示常用光圈值
- **焦距分布**: 柱状图展示常用焦段
- **统计概览**: 总图片数、最多使用的机型/ISO/光圈

### 数据管理
- **数据库存储**: 支持 MySQL 和 SQLite
- **Excel 导出**: 生成带格式的 Excel 报表
- **JSON 导出**: 导出 JSON 格式数据
- **导入历史**: 支持从 JSON 文件导入或从数据库加载历史数据
- **数据下载**: Web 界面一键下载 Excel/JSON 文件

### Web 界面
- **仪表盘**: 扫描控制、结果展示、图表分析一体化
- **目录选择器**: 可视化文件夹浏览（已优化，仅显示目录）
- **设置页面**: 在线配置数据库、导出选项等
- **响应式设计**: 适配不同屏幕尺寸

## 🚀 快速开始

### 1. 配置

首次运行会自动生成 `config.yaml`，也可手动修改：

```yaml
server:
  port: 8080

database:
  enabled: true
  driver: sqlite    # sqlite 或 mysql
  source: exif_data.db
  table: photo_data

scan:
  path: ""
  extensions: [".jpg", ".jpeg", ".png", ".arw", ".tiff"]

excel:
  enabled: true
  output: "scan_results.xlsx"

json:
  enabled: true
  output: "scan_results.json"
```

### 2. 运行

```bash
./ExifScan.exe
# 或指定配置文件
./ExifScan.exe -config my_config.yaml
```

### 3. 使用 Web 界面

打开浏览器访问 `http://localhost:8080`

#### 仪表盘功能
- **选择目录** → 点击"📂 选择"浏览文件夹
- **开始扫描** → 点击"🚀 开始扫描"，完成后自动展示图表和数据
- **加载历史** → 从数据库加载之前保存的扫描记录
- **导入JSON** → 上传 JSON 文件查看图表分析
- **下载结果** → 扫描完成后可一键下载 Excel / JSON

#### 设置页面
- 访问 `http://localhost:8080/settings.html`
- 配置数据库连接、导出选项等

## 🔌 API 接口

| 方法   | 路径                  | 说明                          |
| ------ | --------------------- | ----------------------------- |
| `GET`  | `/api/config`         | 获取当前配置                  |
| `POST` | `/api/config`         | 更新配置                      |
| `POST` | `/api/scan`           | 执行扫描（返回完整结果+统计） |
| `GET`  | `/api/results`        | 从数据库加载历史数据          |
| `POST` | `/api/results/import` | 导入 JSON 文件                |
| `GET`  | `/api/download/excel` | 下载 Excel 文件               |
| `GET`  | `/api/download/json`  | 下载 JSON 文件                |
| `GET`  | `/api/fs/list?path=`  | 浏览目录结构                  |

## 🔨 构建

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

## 🧪 测试

```bash
go test ./... -v
```

## 项目结构

```
ExifScan/
├── cmd/exifscan/       # 入口
│   └── main.go
├── internal/
│   ├── config/         # 配置加载
│   ├── db/             # 数据库操作（MySQL/SQLite）
│   ├── excel/          # Excel 导出
│   ├── model/          # 数据模型
│   ├── scan/           # EXIF 扫描（并发）
│   └── web/            # Web 服务 + 静态资源（嵌入）
│       ├── handlers.go # API 处理器
│       ├── server.go   # 路由注册
│       └── static/     # HTML/CSS（embed 嵌入）
├── config.yaml         # 运行时配置
├── build.bat / build.sh
└── go.mod
```
