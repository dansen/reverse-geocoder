package rgeocoder

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// DataLoader 负责加载数据
type DataLoader struct {
	config  *Config
	verbose bool
}

func NewDataLoader(cfg *Config) *DataLoader { return &DataLoader{config: cfg, verbose: cfg.Verbose} }

var expectedHeader = []string{"lat", "lon", "name", "admin1", "admin2", "cc"}

// LoadFromFile 读取 rg_cities1000.csv (未实现)
func (dl *DataLoader) LoadFromFile(filename string) ([]Coordinate, []Location, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	if dl.verbose {
		fmt.Println("loading dataset:", filename)
	}
	return dl.parseCSV(file)
}

// LoadFromStream 从自定义流读取
func (dl *DataLoader) LoadFromStream(r io.Reader) ([]Coordinate, []Location, error) {
	return dl.parseCSV(r)
}

// ExtractAndProcess 下载+处理 GeoNames (未实现)
func (dl *DataLoader) ExtractAndProcess() ([]Coordinate, []Location, error) {
	// 未来: 下载 + 生成 rg_cities1000.csv
	return nil, nil, errors.New("ExtractAndProcess not implemented")
}

// parseCSV 通用解析
func (dl *DataLoader) parseCSV(r io.Reader) ([]Coordinate, []Location, error) {
	reader := csv.NewReader(bufio.NewReader(r))
	head, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("read header: %w", err)
	}
	if err := dl.validateHeader(head); err != nil {
		return nil, nil, err
	}
	coords := make([]Coordinate, 0, 1024)
	locs := make([]Location, 0, 1024)
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("read record: %w", err)
		}
		if len(rec) < 6 {
			continue
		}
		lat, err1 := strconv.ParseFloat(rec[0], 64)
		lon, err2 := strconv.ParseFloat(rec[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		coords = append(coords, Coordinate{Lat: lat, Lon: lon})
		locs = append(locs, Location{Lat: rec[0], Lon: rec[1], Name: rec[2], Admin1: rec[3], Admin2: rec[4], CC: rec[5]})
	}
	return coords, locs, nil
}

func (dl *DataLoader) validateHeader(head []string) error {
	if len(head) != len(expectedHeader) {
		return fmt.Errorf("unexpected header column count: %d", len(head))
	}
	for i, col := range expectedHeader {
		if head[i] != col {
			return fmt.Errorf("invalid header at %d: got %s want %s", i, head[i], col)
		}
	}
	return nil
}

// Helper to get default dataset path
func datasetPath(dataDir string) string { return filepath.Join(dataDir, "rg_cities1000.csv") }
