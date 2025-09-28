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
	root   *Node
	points []Coordinate
}

// NewKDTree 构建中位数分割KD树
func NewKDTree(points []Coordinate) *KDTree {
	t := &KDTree{points: points}
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
	} // 当前只实现最近一个
	dists := make([]float64, len(coords))
	indices := make([]int, len(coords))
	for i, q := range coords {
		bestIdx := -1
		bestDist := math.MaxFloat64
		searchNN(t.root, q, &bestIdx, &bestDist)
		if bestIdx == -1 { // 空树
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
func searchNN(node *Node, target Coordinate, bestIdx *int, bestDist *float64) {
	if node == nil {
		return
	}
	d := haversine(target.Lat, target.Lon, node.Point.Lat, node.Point.Lon)
	if d < *bestDist {
		*bestDist = d
		*bestIdx = node.Index
	}
	var goLeft bool
	if node.Axis == 0 { // 比较纬度
		goLeft = target.Lat < node.Point.Lat
	} else {
		goLeft = target.Lon < node.Point.Lon
	}
	first := node.Left
	second := node.Right
	if !goLeft {
		first, second = second, first
	}
	searchNN(first, target, bestIdx, bestDist)
	// 判断是否需要访问另一侧
	var axisDist float64
	if node.Axis == 0 {
		axisDist = math.Abs(target.Lat - node.Point.Lat)
	} else {
		axisDist = math.Abs(target.Lon - node.Point.Lon)
	}
	// 粗略用地表距离界限: 如果轴向差转换成近似距离小于当前bestDist才回溯
	// 这里用 111km * 度差 近似
	if axisDist*111.0 < *bestDist {
		searchNN(second, target, bestIdx, bestDist)
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
