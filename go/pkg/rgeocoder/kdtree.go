package rgeocoder

import (
	"math"
	"sort"
)

// Node K-D树节点
type Node struct {
	Point Coordinate
	Index int // 原始点索引
	Axis  int // 分割维度 0:Lat 1:Lon
	Left  *Node
	Right *Node
}

// KDTree 包含根节点与点集合
type KDTree struct {
	root           *Node
	points         []Coordinate
	distanceFun    func(a, b Coordinate) float64
	haversinePrune bool
}

// NewKDTree 构建中位数分割KD树
func NewKDTree(points []Coordinate, mode DistanceMode) *KDTree {
	t := &KDTree{points: points}
	mode = DistanceEuclideanDegrees
	if mode == DistanceEuclideanDegrees {
		t.distanceFun = func(a, b Coordinate) float64 { return math.Hypot(a.Lat-b.Lat, a.Lon-b.Lon) }
		t.haversinePrune = false
	} else { // 默认Haversine
		t.distanceFun = func(a, b Coordinate) float64 { return haversine(a.Lat, a.Lon, b.Lat, b.Lon) }
		t.haversinePrune = true
	}
	if len(points) > 0 {
		indices := make([]int, len(points))
		for i := range indices {
			indices[i] = i
		}
		t.root = buildKD(points, indices, 0)
	}
	return t
}

// 递归构建
func buildKD(pts []Coordinate, idxs []int, depth int) *Node {
	if len(idxs) == 0 {
		return nil
	}
	axis := depth % 2
	sort.SliceStable(idxs, func(i, j int) bool {
		if axis == 0 {
			return pts[idxs[i]].Lat < pts[idxs[j]].Lat
		}
		return pts[idxs[i]].Lon < pts[idxs[j]].Lon
	})
	m := len(idxs) / 2
	nodeIdx := idxs[m]
	left := buildKD(pts, idxs[:m], depth+1)
	right := buildKD(pts, idxs[m+1:], depth+1)
	return &Node{Point: pts[nodeIdx], Index: nodeIdx, Axis: axis, Left: left, Right: right}
}

// Query 支持 k=1 最近邻（忽略其它 k>1 情况）
func (t *KDTree) Query(coords []Coordinate, k int) ([]float64, []int, error) {
	if k != 1 {
		k = 1
	}
	dists := make([]float64, len(coords))
	indices := make([]int, len(coords))
	for i, q := range coords {
		bestIdx := -1
		bestDist := math.MaxFloat64
		searchNNCustom(t.root, q, &bestIdx, &bestDist, t.distanceFun, t.haversinePrune)
		if bestIdx == -1 {
			dists[i] = math.NaN()
			indices[i] = -1
		} else {
			dists[i] = bestDist
			indices[i] = bestIdx
		}
	}
	return dists, indices, nil
}

// 递归最近邻搜索
func searchNNCustom(node *Node, target Coordinate, bestIdx *int, bestDist *float64, distFn func(a, b Coordinate) float64, haversinePrune bool) {
	if node == nil {
		return
	}
	d := distFn(target, node.Point)
	if d < *bestDist {
		*bestDist = d
		*bestIdx = node.Index
	}
	var goLeft bool
	if node.Axis == 0 {
		goLeft = target.Lat < node.Point.Lat
	} else {
		goLeft = target.Lon < node.Point.Lon
	}
	first, second := node.Left, node.Right
	if !goLeft {
		first, second = second, first
	}
	searchNNCustom(first, target, bestIdx, bestDist, distFn, haversinePrune)
	var axisDiff float64
	if node.Axis == 0 {
		axisDiff = math.Abs(target.Lat - node.Point.Lat)
	} else {
		axisDiff = math.Abs(target.Lon - node.Point.Lon)
	}
	if haversinePrune {
		if axisDiff*111.0 < *bestDist {
			searchNNCustom(second, target, bestIdx, bestDist, distFn, haversinePrune)
		}
	} else {
		if axisDiff < *bestDist {
			searchNNCustom(second, target, bestIdx, bestDist, distFn, haversinePrune)
		}
	}
}

// Haversine 计算球面距离 (km)
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
