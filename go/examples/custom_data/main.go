package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

// 演示如何用自定义数据流创建实例
func main() {
	csvData := "lat,lon,name,admin1,admin2,cc\n37.78674,-122.39222,SampleCity,Region,Sub,US\n"
	r := bytes.NewBufferString(csvData)
	rg, err := rgeocoder.NewRGeocoderWithStream(r, rgeocoder.WithVerbose(true))
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	loc, err := rg.QuerySingle(rgeocoder.Coordinate{Lat: 37.78674, Lon: -122.39222})
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	fmt.Println("结果:", loc.Name, loc.CC)
}
