package rgeocoder

import (
	"errors"
	"math"
	"os"
	"path/filepath"
)

const (
	WGS84MajorAxis           = 6378.137
	WGS84EccentricitySquared = 0.00669437999014
	EarthRadius              = 6371.0
)

// ValidateCoordinate 检查单个坐标
func ValidateCoordinate(c Coordinate) error {
	if c.Lat < -90 || c.Lat > 90 || c.Lon < -180 || c.Lon > 180 {
		return errors.New("coordinate out of range")
	}
	return nil
}

// ValidateCoordinates 批量校验
func ValidateCoordinates(cs []Coordinate) error {
	for _, c := range cs {
		if err := ValidateCoordinate(c); err != nil {
			return err
		}
	}
	return nil
}

// GeodeticToECEF 简化版本（占位）
func GeodeticToECEF(coords []Coordinate) [][]float64 {
	out := make([][][]float64, 0)
	_ = out
	return [][]float64{}
}

// HaversineDistance 供外部使用
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(x float64) float64 { return x * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	la1 := toRad(lat1)
	la2 := toRad(lat2)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

// GetDataDir 返回默认数据目录
func GetDataDir() string { return filepath.Join(".", "go", "data") }

// EnsureDir 确保目录存在
func EnsureDir(dir string) error { return os.MkdirAll(dir, 0o755) }
