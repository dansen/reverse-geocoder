package rgeocoder

import (
	"context"
	"sync"
)

// KDTreeMP 并发包装（当前直接调用单线程实现）
type KDTreeMP struct {
	base    *KDTree
	workers int
}

// NewKDTreeMP 创建多线程版本（占位）
func NewKDTreeMP(points []Coordinate, workers int, mode DistanceMode) *KDTreeMP {
	if workers <= 0 {
		workers = 4
	}
	return &KDTreeMP{base: NewKDTree(points, mode), workers: workers}
}

// Query 并发查询（暂时顺序调用）
func (t *KDTreeMP) Query(coords []Coordinate, k int) ([]float64, []int, error) {
	// 简单拆分任务，但由于base是线性扫描，收益有限
	if len(coords) < 2 || t.workers <= 1 {
		return t.base.Query(coords, k)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type part struct {
		idx   int
		coord Coordinate
	}
	jobs := make(chan part)
	var wg sync.WaitGroup
	results := make([]struct {
		d float64
		i int
	}, len(coords))

	worker := func() {
		defer wg.Done()
		for p := range jobs {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, inds, _ := t.base.Query([]Coordinate{p.coord}, 1)
			results[p.idx] = struct {
				d float64
				i int
			}{d: 0, i: inds[0]}
		}
	}

	wg.Add(t.workers)
	for w := 0; w < t.workers; w++ {
		go worker()
	}
	for i, c := range coords {
		jobs <- part{idx: i, coord: c}
	}
	close(jobs)
	wg.Wait()

	dists := make([]float64, len(coords))
	indices := make([]int, len(coords))
	for i, r := range results {
		dists[i] = r.d
		indices[i] = r.i
	}
	return dists, indices, nil
}
