package rgeocoder

import "math"

// Node 简化K-D树节点（占位）
type Node struct {
	Point Coordinate
	Index int
	Left  *Node
	Right *Node
}

// KDTree 简化实现（当前只线性扫描）
type KDTree struct {
	points []Coordinate
}

// NewKDTree 创建K-D树（暂不真正构建）
func NewKDTree(points []Coordinate) *KDTree { return &KDTree{points: points} }

// Query 最近邻（k=1）线性扫描实现
func (t *KDTree) Query(coords []Coordinate, k int) ([]float64, []int, error) {
	dists := make([]float64, len(coords))
	indices := make([]int, len(coords))
	for i, c := range coords {
		bestDist := math.MaxFloat64
		bestIdx := -1
		for idx, p := range t.points {
			d := haversine(c.Lat, c.Lon, p.Lat, p.Lon)
			if d < bestDist {
				bestDist = d
				bestIdx = idx
			}
		}
		if bestIdx == -1 { // 空数据集
			bestIdx = 0
			bestDist = math.NaN()
		}
		dists[i] = bestDist
		indices[i] = bestIdx
	}
	return dists, indices, nil
}

// 简单Haversine
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	toRad := func(x float64) float64 { return x * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	la1 := toRad(lat1)
	la2 := toRad(lat2)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
