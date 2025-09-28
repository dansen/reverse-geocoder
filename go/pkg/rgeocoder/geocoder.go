package rgeocoder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// KDTreeInterface 允许不同实现（单线程 / 多线程）
type KDTreeInterface interface {
	Query(coords []Coordinate, k int) ([]float64, []int, error)
}

// RGeocoder 主结构体
type RGeocoder struct {
	mode      QueryMode
	verbose   bool
	tree      KDTreeInterface
	locations []Location
	mu        sync.RWMutex
	config    *Config
}

// Option 函数式配置
type Option func(*Config)

// WithMode 设置模式
func WithMode(mode QueryMode) Option { return func(c *Config) { c.Mode = mode } }

// WithVerbose 设置日志
func WithVerbose(v bool) Option { return func(c *Config) { c.Verbose = v } }

// WithDataDir 设置数据目录
func WithDataDir(dir string) Option { return func(c *Config) { c.DataDir = dir } }

// WithMaxWorkers 设置并发
func WithMaxWorkers(n int) Option { return func(c *Config) { c.MaxWorkers = n } }

// WithDistanceMode 设置距离模式
func WithDistanceMode(m DistanceMode) Option { return func(c *Config) { c.DistanceMode = m } }

// applyOptions 应用默认与用户选项
func applyOptions(opts []Option) *Config {
	cfg := &Config{
		Mode:         MultiThreaded,
		Verbose:      false,
		DataDir:      filepath.Join(".", "data"),
		DownloadURLs: DefaultURLs,
		MaxWorkers:   0,
		CacheEnabled: true,
		DistanceMode: DistanceHaversine,
	}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 4
	}
	return cfg
}

// NewRGeocoder 创建实例（自动加载或下载数据的逻辑待实现）
func NewRGeocoder(opts ...Option) (*RGeocoder, error) {
	cfg := applyOptions(opts)
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	// 加载数据 (允许空数据集，而不是 panic，以便测试和首次运行)
	citiesFile := filepath.Join(cfg.DataDir, "rg_cities1000.csv")
	if cfg.Verbose {
		fmt.Printf("checking data file: %s\n", citiesFile)
	}

	var coords []Coordinate
	var locs []Location
	if _, errStat := os.Stat(citiesFile); errors.Is(errStat, os.ErrNotExist) {
		if cfg.Verbose {
			fmt.Println("dataset not found, starting with empty dataset (place file at:", citiesFile, ")")
		}
	} else {
		loader := NewDataLoader(cfg)
		c, l, err := loader.LoadFromFile(citiesFile)
		if err != nil {
			if cfg.Verbose {
				fmt.Println("failed to load dataset:", err)
			}
		} else {
			coords, locs = c, l
			if cfg.Verbose {
				fmt.Printf("loaded %d locations\n", len(locs))
			}
		}
	}

	// 构建KD树，传入距离模式
	var tree KDTreeInterface
	if cfg.Mode == SingleThreaded {
		tree = NewKDTree(coords, cfg.DistanceMode)
	} else {
		tree = NewKDTreeMP(coords, cfg.MaxWorkers, cfg.DistanceMode)
	}

	return &RGeocoder{mode: cfg.Mode, verbose: cfg.Verbose, tree: tree, locations: locs, config: cfg}, nil
}

// NewRGeocoderWithStream 使用内存流初始化
func NewRGeocoderWithStream(stream io.Reader, opts ...Option) (*RGeocoder, error) {
	cfg := applyOptions(opts)
	loader := NewDataLoader(cfg)
	coords, locs, err := loader.LoadFromStream(stream)
	if err != nil {
		return nil, err
	}
	var tree KDTreeInterface
	if cfg.Mode == SingleThreaded {
		tree = NewKDTree(coords, cfg.DistanceMode)
	} else {
		tree = NewKDTreeMP(coords, cfg.MaxWorkers, cfg.DistanceMode)
	}
	return &RGeocoder{mode: cfg.Mode, verbose: cfg.Verbose, tree: tree, locations: locs, config: cfg}, nil
}

// Query 批量查询
func (rg *RGeocoder) Query(coordinates []Coordinate) ([]Location, error) {
	if len(coordinates) == 0 {
		return nil, fmt.Errorf("no coordinates provided")
	}
	// 验证
	for _, c := range coordinates {
		if c.Lat < -90 || c.Lat > 90 || c.Lon < -180 || c.Lon > 180 {
			return nil, fmt.Errorf("invalid coordinate: %+v", c)
		}
	}
	_, indices, err := rg.tree.Query(coordinates, 1)
	if err != nil {
		return nil, err
	}
	results := make([]Location, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(rg.locations) {
			results = append(results, rg.locations[idx])
		} else {
			results = append(results, Location{})
		}
	}
	return results, nil
}

// QuerySingle 单个查询
func (rg *RGeocoder) QuerySingle(c Coordinate) (Location, error) {
	locs, err := rg.Query([]Coordinate{c})
	if err != nil || len(locs) == 0 {
		return Location{}, err
	}
	return locs[0], nil
}

// Close 释放资源（当前无状态）
func (rg *RGeocoder) Close() error { return nil }

// Get 便捷函数
func Get(coord Coordinate, opts ...Option) (Location, error) {
	rg, err := NewRGeocoder(opts...)
	if err != nil {
		return Location{}, err
	}
	defer rg.Close()
	return rg.QuerySingle(coord)
}

// Search 便捷批量函数
func Search(coords []Coordinate, opts ...Option) ([]Location, error) {
	rg, err := NewRGeocoder(opts...)
	if err != nil {
		return nil, err
	}
	defer rg.Close()
	return rg.Query(coords)
}
