package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
	mode := flag.Int("mode", 2, "查询模式: 1=单线程 2=多线程")
	verbose := flag.Bool("verbose", false, "是否输出详细日志")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: %s [--mode 1|2] [--verbose] <lat> <lon>\n", os.Args[0])
		os.Exit(1)
	}

	lat, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		log.Fatalf("纬度解析失败: %v", err)
	}
	lon, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		log.Fatalf("经度解析失败: %v", err)
	}

	rg, err := rgeocoder.NewRGeocoder(
		rgeocoder.WithMode(rgeocoder.QueryMode(*mode)),
		rgeocoder.WithVerbose(*verbose),
	)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer rg.Close()

	loc, err := rg.QuerySingle(rgeocoder.Coordinate{Lat: lat, Lon: lon})
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	fmt.Printf("Result => name=%s admin1=%s admin2=%s cc=%s (lat=%s lon=%s)\n", loc.Name, loc.Admin1, loc.Admin2, loc.CC, loc.Lat, loc.Lon)
}
