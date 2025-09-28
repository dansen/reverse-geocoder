package main

import (
	"fmt"
	"log"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
	coords := []rgeocoder.Coordinate{
		{Lat: 51.5214588, Lon: -0.1729636},
		{Lat: 9.936033, Lon: 76.259952},
		{Lat: 37.38605, Lon: -122.08385},
	}
	locs, err := rgeocoder.Search(coords, rgeocoder.WithVerbose(true))
	if err != nil {
		log.Fatalf("批量查询失败: %v", err)
	}
	for i, l := range locs {
		fmt.Printf("%d => %s %s %s (%s,%s)\n", i, l.Name, l.Admin1, l.CC, l.Lat, l.Lon)
	}
}
