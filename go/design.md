# Reverse Geocoder - Go版本设计文档

## 项目概述

基于Python版本的reverse-geocoder库，设计一个纯Go实现的离线反向地理编码库。该库能够根据经纬度坐标快速查找最近的城市信息，支持单线程和多线程模式。

## 核心功能

- 离线反向地理编码（根据经纬度查找城市）
- 支持单个坐标和批量坐标查询
- 使用K-D树算法进行快速最近邻搜索
- 支持单线程和多线程并发查询
- 自动下载和处理GeoNames数据
- 支持自定义数据源

## 目录结构

```
go/
├── README.md                    # 项目说明文档
├── go.mod                       # Go模块定义
├── go.sum                       # 依赖校验和
├── cmd/
│   └── rgeocoder/
│       └── main.go              # 命令行工具入口
├── pkg/
│   └── rgeocoder/
│       ├── geocoder.go          # 主要的反向地理编码器
│       ├── kdtree.go            # K-D树实现
│       ├── kdtree_mp.go         # 多线程K-D树实现
│       ├── data.go              # 数据结构定义
│       ├── loader.go            # 数据加载器
│       ├── downloader.go        # GeoNames数据下载器
│       ├── processor.go         # 数据处理器
│       └── utils.go             # 工具函数
├── internal/
│   └── config/
│       └── config.go            # 配置管理
├── data/
│   └── .gitkeep                 # 数据目录占位符
├── examples/
│   ├── basic/
│   │   └── main.go              # 基本使用示例
│   ├── batch/
│   │   └── main.go              # 批量查询示例
│   └── custom_data/
│       └── main.go              # 自定义数据源示例
└── tests/
    ├── geocoder_test.go         # 主要功能测试
    ├── kdtree_test.go           # K-D树测试
    ├── benchmarks_test.go       # 性能基准测试
    └── testdata/
        └── sample_cities.csv    # 测试数据
```

## 核心文件设计

### 1. pkg/rgeocoder/data.go - 数据结构定义

```go
package rgeocoder

import "time"

// Coordinate 表示地理坐标
type Coordinate struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}

// Location 表示地理位置信息
type Location struct {
    Lat    string `json:"lat" csv:"lat"`
    Lon    string `json:"lon" csv:"lon"`
    Name   string `json:"name" csv:"name"`
    Admin1 string `json:"admin1" csv:"admin1"`
    Admin2 string `json:"admin2" csv:"admin2"`
    CC     string `json:"cc" csv:"cc"`
}

// GeoNamesRecord 表示GeoNames数据记录
type GeoNamesRecord struct {
    GeoNameID        int       `csv:"geonameid"`
    Name             string    `csv:"name"`
    ASCIIName        string    `csv:"asciiname"`
    AlternateNames   string    `csv:"alternatenames"`
    Latitude         float64   `csv:"latitude"`
    Longitude        float64   `csv:"longitude"`
    FeatureClass     string    `csv:"feature_class"`
    FeatureCode      string    `csv:"feature_code"`
    CountryCode      string    `csv:"country_code"`
    CC2              string    `csv:"cc2"`
    Admin1Code       string    `csv:"admin1_code"`
    Admin2Code       string    `csv:"admin2_code"`
    Admin3Code       string    `csv:"admin3_code"`
    Admin4Code       string    `csv:"admin4_code"`
    Population       int       `csv:"population"`
    Elevation        int       `csv:"elevation"`
    DEM              int       `csv:"dem"`
    Timezone         string    `csv:"timezone"`
    ModificationDate time.Time `csv:"modification_date"`
}

// AdminRecord 表示行政区划记录
type AdminRecord struct {
    ConcatCodes string `csv:"concat_codes"`
    Name        string `csv:"name"`
    ASCIIName   string `csv:"ascii_name"`
    GeoNameID   int    `csv:"geonameid"`
}

// QueryMode 查询模式
type QueryMode int

const (
    SingleThreaded QueryMode = 1 // 单线程模式
    MultiThreaded  QueryMode = 2 // 多线程模式（默认）
)

// Config 配置结构
type Config struct {
    Mode           QueryMode
    Verbose        bool
    DataDir        string
    DownloadURLs   URLs
    MaxWorkers     int
    CacheEnabled   bool
}

// URLs GeoNames数据下载地址
type URLs struct {
    BaseURL        string
    Cities1000     string
    Admin1Codes    string
    Admin2Codes    string
}
```

### 2. pkg/rgeocoder/geocoder.go - 主要地理编码器

```go
package rgeocoder

import (
    "fmt"
    "io"
    "sync"
)

// RGeocoder 反向地理编码器
type RGeocoder struct {
    mode      QueryMode
    verbose   bool
    tree      KDTreeInterface
    locations []Location
    mu        sync.RWMutex
}

// KDTreeInterface K-D树接口
type KDTreeInterface interface {
    Query(coordinates []Coordinate, k int) (distances []float64, indices []int, error)
}

// NewRGeocoder 创建新的反向地理编码器实例
func NewRGeocoder(options ...Option) (*RGeocoder, error)

// NewRGeocoderWithStream 使用自定义数据流创建编码器
func NewRGeocoderWithStream(stream io.Reader, options ...Option) (*RGeocoder, error)

// Query 查询坐标对应的位置信息
func (rg *RGeocoder) Query(coordinates []Coordinate) ([]Location, error)

// QuerySingle 查询单个坐标
func (rg *RGeocoder) QuerySingle(coord Coordinate) (Location, error)

// Close 释放资源
func (rg *RGeocoder) Close() error

// Option 配置选项函数类型
type Option func(*Config)

// WithMode 设置查询模式
func WithMode(mode QueryMode) Option

// WithVerbose 设置详细输出
func WithVerbose(verbose bool) Option

// WithDataDir 设置数据目录
func WithDataDir(dataDir string) Option

// WithMaxWorkers 设置最大工作协程数
func WithMaxWorkers(workers int) Option

// 全局便捷函数
func Get(coord Coordinate, options ...Option) (Location, error)
func Search(coords []Coordinate, options ...Option) ([]Location, error)
```

### 3. pkg/rgeocoder/kdtree.go - K-D树实现

```go
package rgeocoder

// Node K-D树节点
type Node struct {
    Point     Coordinate
    Index     int
    Dimension int
    Left      *Node
    Right     *Node
}

// KDTree K-D树结构
type KDTree struct {
    root      *Node
    points    []Coordinate
    dimension int
}

// NewKDTree 创建新的K-D树
func NewKDTree(points []Coordinate) *KDTree

// Query 查询最近邻
func (t *KDTree) Query(coordinates []Coordinate, k int) (distances []float64, indices []int, error)

// Build 构建K-D树
func (t *KDTree) Build(points []Coordinate, indices []int, depth int) *Node

// SearchNearest 搜索最近的点
func (t *KDTree) SearchNearest(target Coordinate, node *Node, best *BestMatch, depth int)

// BestMatch 最佳匹配结果
type BestMatch struct {
    Index    int
    Distance float64
}
```

### 4. pkg/rgeocoder/kdtree_mp.go - 多线程K-D树

```go
package rgeocoder

import (
    "context"
    "runtime"
    "sync"
)

// KDTreeMP 多线程K-D树
type KDTreeMP struct {
    *KDTree
    workers    int
    workerPool chan struct{}
}

// NewKDTreeMP 创建多线程K-D树
func NewKDTreeMP(points []Coordinate, workers int) *KDTreeMP

// Query 并行查询
func (t *KDTreeMP) Query(coordinates []Coordinate, k int) (distances []float64, indices []int, error)

// queryWorker 工作协程
func (t *KDTreeMP) queryWorker(ctx context.Context, jobs <-chan QueryJob, results chan<- QueryResult, wg *sync.WaitGroup)

// QueryJob 查询任务
type QueryJob struct {
    Coordinates []Coordinate
    StartIndex  int
    K          int
}

// QueryResult 查询结果
type QueryResult struct {
    Distances  []float64
    Indices    []int
    StartIndex int
    Error      error
}
```

### 5. pkg/rgeocoder/loader.go - 数据加载器

```go
package rgeocoder

import (
    "encoding/csv"
    "io"
    "os"
)

// DataLoader 数据加载器
type DataLoader struct {
    config  *Config
    verbose bool
}

// NewDataLoader 创建数据加载器
func NewDataLoader(config *Config) *DataLoader

// LoadFromFile 从文件加载数据
func (dl *DataLoader) LoadFromFile(filename string) ([]Coordinate, []Location, error)

// LoadFromStream 从流加载数据
func (dl *DataLoader) LoadFromStream(stream io.Reader) ([]Coordinate, []Location, error)

// ExtractAndProcess 提取并处理GeoNames数据
func (dl *DataLoader) ExtractAndProcess() ([]Coordinate, []Location, error)

// parseCSVRecord 解析CSV记录
func (dl *DataLoader) parseCSVRecord(record []string) (Location, error)

// validateHeader 验证CSV头部
func (dl *DataLoader) validateHeader(header []string) error
```

### 6. pkg/rgeocoder/downloader.go - 数据下载器

```go
package rgeocoder

import (
    "archive/zip"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

// Downloader GeoNames数据下载器
type Downloader struct {
    config    *Config
    client    *http.Client
    verbose   bool
}

// NewDownloader 创建下载器
func NewDownloader(config *Config) *Downloader

// DownloadRequired 下载必需的文件
func (d *Downloader) DownloadRequired() error

// downloadFile 下载单个文件
func (d *Downloader) downloadFile(url, filename string) error

// extractZip 解压ZIP文件
func (d *Downloader) extractZip(src, dest string) error

// fileExists 检查文件是否存在
func (d *Downloader) fileExists(filename string) bool
```

### 7. pkg/rgeocoder/processor.go - 数据处理器

```go
package rgeocoder

import (
    "encoding/csv"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

// DataProcessor 数据处理器
type DataProcessor struct {
    config      *Config
    verbose     bool
    admin1Map   map[string]string
    admin2Map   map[string]string
}

// NewDataProcessor 创建数据处理器
func NewDataProcessor(config *Config) *DataProcessor

// ProcessGeoNamesData 处理GeoNames原始数据
func (dp *DataProcessor) ProcessGeoNamesData() error

// loadAdminCodes 加载行政区划代码
func (dp *DataProcessor) loadAdminCodes() error

// loadAdmin1Codes 加载一级行政区划
func (dp *DataProcessor) loadAdmin1Codes() error

// loadAdmin2Codes 加载二级行政区划
func (dp *DataProcessor) loadAdmin2Codes() error

// processCitiesFile 处理城市文件
func (dp *DataProcessor) processCitiesFile() error

// parseGeoNamesRecord 解析GeoNames记录
func (dp *DataProcessor) parseGeoNamesRecord(record []string) (*GeoNamesRecord, error)

// convertToLocation 转换为Location结构
func (dp *DataProcessor) convertToLocation(gnRecord *GeoNamesRecord) Location
```

### 8. pkg/rgeocoder/utils.go - 工具函数

```go
package rgeocoder

import (
    "math"
)

const (
    // WGS84长半轴 (千米)
    WGS84MajorAxis = 6378.137
    // WGS84偏心率的平方
    WGS84EccentricitySquared = 0.00669437999014
    // 地球半径 (千米)
    EarthRadius = 6371.0
)

// Distance 计算两点间的距离
func Distance(p1, p2 Coordinate) float64

// HaversineDistance 使用Haversine公式计算距离
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64

// DegreesToRadians 角度转弧度
func DegreesToRadians(degrees float64) float64

// RadiansToDegrees 弧度转角度
func RadiansToDegrees(radians float64) float64

// GeodeticToECEF 大地坐标转ECEF坐标
func GeodeticToECEF(coords []Coordinate) [][]float64

// ValidateCoordinate 验证坐标有效性
func ValidateCoordinate(coord Coordinate) error

// ValidateCoordinates 批量验证坐标
func ValidateCoordinates(coords []Coordinate) error

// GetDataDir 获取数据目录路径
func GetDataDir() string

// EnsureDir 确保目录存在
func EnsureDir(dir string) error
```

### 9. cmd/rgeocoder/main.go - 命令行工具

```go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"
    
    "github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
    var (
        mode     = flag.Int("mode", 2, "查询模式: 1=单线程, 2=多线程")
        verbose  = flag.Bool("verbose", false, "详细输出")
        format   = flag.String("format", "json", "输出格式: json, csv, table")
        batch    = flag.String("batch", "", "批量查询文件路径")
    )
    flag.Parse()

    if len(flag.Args()) < 2 && *batch == "" {
        fmt.Fprintf(os.Stderr, "用法: %s [选项] <纬度> <经度>\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "或者: %s [选项] -batch <文件路径>\n", os.Args[0])
        flag.PrintDefaults()
        os.Exit(1)
    }

    // 创建地理编码器
    geocoder, err := rgeocoder.NewRGeocoder(
        rgeocoder.WithMode(rgeocoder.QueryMode(*mode)),
        rgeocoder.WithVerbose(*verbose),
    )
    if err != nil {
        log.Fatalf("初始化地理编码器失败: %v", err)
    }
    defer geocoder.Close()

    if *batch != "" {
        // 批量处理
        handleBatchQuery(geocoder, *batch, *format)
    } else {
        // 单个查询
        handleSingleQuery(geocoder, flag.Args(), *format)
    }
}

func handleSingleQuery(geocoder *rgeocoder.RGeocoder, args []string, format string) {
    // 实现单个查询逻辑
}

func handleBatchQuery(geocoder *rgeocoder.RGeocoder, filename, format string) {
    // 实现批量查询逻辑
}
```

### 10. examples/ - 使用示例

#### examples/basic/main.go
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
    // 基本单个坐标查询
    coord := rgeocoder.Coordinate{Lat: 37.78674, Lon: -122.39222}
    location, err := rgeocoder.Get(coord, rgeocoder.WithVerbose(true))
    if err != nil {
        log.Fatalf("查询失败: %v", err)
    }
    
    fmt.Printf("坐标 (%.5f, %.5f) 对应的城市: %s, %s, %s\n", 
        coord.Lat, coord.Lon, location.Name, location.Admin1, location.CC)
}
```

#### examples/batch/main.go
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
    // 批量坐标查询
    coords := []rgeocoder.Coordinate{
        {Lat: 51.5214588, Lon: -0.1729636}, // 伦敦
        {Lat: 9.936033, Lon: 76.259952},    // 印度某地
        {Lat: 37.38605, Lon: -122.08385},   // 美国某地
    }
    
    locations, err := rgeocoder.Search(coords, 
        rgeocoder.WithMode(rgeocoder.MultiThreaded),
        rgeocoder.WithVerbose(true))
    if err != nil {
        log.Fatalf("批量查询失败: %v", err)
    }
    
    for i, location := range locations {
        fmt.Printf("坐标 %d: %s, %s, %s\n", i+1, 
            location.Name, location.Admin1, location.CC)
    }
}
```

## 核心特性

### 1. 性能特性
- 使用K-D树算法实现O(log n)的查询复杂度
- 支持多协程并行查询
- 内存高效的数据结构设计
- 可配置的工作协程池

### 2. 数据管理
- 自动下载GeoNames数据
- 智能缓存管理
- 支持自定义数据源
- 数据完整性验证

### 3. 易用性
- 简洁的API设计
- 丰富的配置选项
- 详细的错误处理
- 完整的使用示例

### 4. 扩展性
- 模块化架构设计
- 可插拔的组件
- 支持自定义距离算法
- 支持多种输出格式

## 依赖管理

### go.mod
```go
module github.com/your-username/reverse-geocoder-go

go 1.21

require (
    github.com/stretchr/testify v1.8.4
    gopkg.in/yaml.v3 v3.0.1
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
)
```

## 测试策略

1. **单元测试**: 覆盖所有核心功能模块
2. **集成测试**: 测试完整的查询流程
3. **性能测试**: 基准测试和压力测试
4. **并发测试**: 多协程安全性测试
5. **边界测试**: 异常输入和边界条件测试

## 部署和分发

1. **命令行工具**: 可独立运行的二进制文件
2. **Go模块**: 可作为库被其他项目引用
3. **Docker镜像**: 容器化部署支持
4. **GitHub Releases**: 自动化构建和发布

## 兼容性说明

- **Go版本**: 要求Go 1.18+（使用泛型特性）
- **平台支持**: Windows, Linux, macOS
- **数据兼容**: 与Python版本使用相同的GeoNames数据格式
- **API兼容**: 提供与Python版本类似的接口设计

这个设计确保了Go版本既保持了原Python版本的核心功能，又充分利用了Go语言的性能和并发特性。