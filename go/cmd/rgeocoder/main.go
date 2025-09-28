package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func main() {
	mode := flag.Int("mode", 2, "查询模式: 1=单线程 2=多线程")
	verbose := flag.Bool("verbose", false, "是否输出详细日志")
	httpAddr := flag.String("http", "8080", "HTTP监听地址(例如 :8080，留空则执行单次查询模式，默认8080)")
	flag.Parse()

	rg, err := rgeocoder.NewRGeocoder(
		rgeocoder.WithMode(rgeocoder.QueryMode(*mode)),
		rgeocoder.WithVerbose(*verbose),
	)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer rg.Close()

	if *httpAddr != "" {
		addr := *httpAddr
		// 允许用户输入 "8080" 或 "0.0.0.0:8080" 或 ":8080"
		if !strings.Contains(addr, ":") {
			addr = ":" + addr
		} else if strings.HasPrefix(addr, ":") && len(addr) == 1 { // 防止传入仅":"
			addr = ":8080"
		}
		if err := runHTTPServer(rg, addr); err != nil {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
		return
	}

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: %s [--mode 1|2] [--verbose] [--http :8080] <lat> <lon>\n", os.Args[0])
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
	loc, err := rg.QuerySingle(rgeocoder.Coordinate{Lat: lat, Lon: lon})
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	fmt.Printf("Result => name=%s admin1=%s admin2=%s cc=%s (lat=%s lon=%s)\n", loc.Name, loc.Admin1, loc.Admin2, loc.CC, loc.Lat, loc.Lon)
}
