package main

import (
	"fmt"
	"log"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
	coord := rgeocoder.Coordinate{Lat: 37.78674, Lon: -122.39222}
	loc, err := rgeocoder.Get(coord, rgeocoder.WithVerbose(true))
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	fmt.Printf("结果: %s %s %s (%s,%s)\n", loc.Name, loc.Admin1, loc.CC, loc.Lat, loc.Lon)
}
