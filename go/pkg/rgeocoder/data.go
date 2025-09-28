package rgeocoder

import "time"

// Coordinate 表示地理坐标
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Location 表示地理位置信息（与Python版本rg_cities1000.csv列一致）
type Location struct {
	Lat    string `json:"lat" csv:"lat"`
	Lon    string `json:"lon" csv:"lon"`
	Name   string `json:"name" csv:"name"`
	Admin1 string `json:"admin1" csv:"admin1"`
	Admin2 string `json:"admin2" csv:"admin2"`
	CC     string `json:"cc" csv:"cc"`
}

// GeoNamesRecord 原始GeoNames城市记录（只保留需要的字段）
type GeoNamesRecord struct {
	GeoNameID        int       `csv:"geonameid"`
	ASCIIName        string    `csv:"asciiname"`
	Latitude         float64   `csv:"latitude"`
	Longitude        float64   `csv:"longitude"`
	CountryCode      string    `csv:"country_code"`
	Admin1Code       string    `csv:"admin1_code"`
	Admin2Code       string    `csv:"admin2_code"`
	Population       int       `csv:"population"`
	ModificationDate time.Time `csv:"modification_date"`
}

// AdminRecord 行政区划编码映射
type AdminRecord struct {
	ConcatCodes string `csv:"concat_codes"`
	Name        string `csv:"name"`
	ASCIIName   string `csv:"ascii_name"`
	GeoNameID   int    `csv:"geonameid"`
}

// QueryMode 查询模式
type QueryMode int

const (
	SingleThreaded QueryMode = 1
	MultiThreaded  QueryMode = 2
)

// Config 全局配置
type Config struct {
	Mode         QueryMode
	Verbose      bool
	DataDir      string
	DownloadURLs URLs
	MaxWorkers   int
	CacheEnabled bool
}

// URLs GeoNames数据下载URL集合
type URLs struct {
	BaseURL     string
	Cities1000  string
	Admin1Codes string
	Admin2Codes string
}

// 默认下载URL
var DefaultURLs = URLs{
	BaseURL:     "http://download.geonames.org/export/dump/",
	Cities1000:  "cities1000.zip",
	Admin1Codes: "admin1CodesASCII.txt",
	Admin2Codes: "admin2Codes.txt",
}
