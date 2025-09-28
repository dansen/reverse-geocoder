package rgeocoder

import (
	"bufio"
	"encoding/csv"
	"io"
)

// DataLoader 负责加载数据
type DataLoader struct {
	config  *Config
	verbose bool
}

func NewDataLoader(cfg *Config) *DataLoader { return &DataLoader{config: cfg, verbose: cfg.Verbose} }

// LoadFromFile 读取 rg_cities1000.csv (未实现)
func (dl *DataLoader) LoadFromFile(filename string) ([]Coordinate, []Location, error) {
	// TODO: 实现真实解析
	return []Coordinate{}, []Location{}, nil
}

// LoadFromStream 从自定义流读取
func (dl *DataLoader) LoadFromStream(r io.Reader) ([]Coordinate, []Location, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	// TODO: 校验表头
	_, _ = reader.Read()
	coords := []Coordinate{}
	locs := []Location{}
	return coords, locs, nil
}

// ExtractAndProcess 下载+处理 GeoNames (未实现)
func (dl *DataLoader) ExtractAndProcess() ([]Coordinate, []Location, error) {
	return []Coordinate{}, []Location{}, nil
}
