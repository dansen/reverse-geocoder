package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

// apiServer 封装 HTTP 逻辑
type apiServer struct {
	geo *rgeocoder.RGeocoder
}

func newAPIServer(geo *rgeocoder.RGeocoder) *apiServer { return &apiServer{geo: geo} }

func (s *apiServer) register(r *gin.Engine) {
	r.GET("/health", s.health)
	r.GET("/reverse", s.reverse)
	r.POST("/batch", s.batch)
}

// 统一响应结构
type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func respond(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, apiResponse{Code: code, Message: message, Data: data})
}

func respondError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, apiResponse{Code: code, Message: message, Data: nil})
}

func (s *apiServer) health(c *gin.Context) {
	respond(c, 0, "ok", gin.H{"time": time.Now().UTC()})
}

// /reverse?lat=..&lon=..
func (s *apiServer) reverse(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	if latStr == "" || lonStr == "" {
		respondError(c, 40001, "missing lat or lon")
		return
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		respondError(c, 40002, "invalid lat or lon")
		return
	}
	loc, err := s.geo.QuerySingle(rgeocoder.Coordinate{Lat: lat, Lon: lon})
	if err != nil {
		respondError(c, 50001, err.Error())
		return
	}
	respond(c, 0, "success", loc)
}

type batchRequest struct {
	Points []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"points"`
}

// batchResponse 去掉外层自定义结构，直接放进 data

func (s *apiServer) batch(c *gin.Context) {
	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, 40010, "invalid body")
		return
	}
	coords := make([]rgeocoder.Coordinate, 0, len(req.Points))
	for _, p := range req.Points {
		coords = append(coords, rgeocoder.Coordinate{Lat: p.Lat, Lon: p.Lon})
	}
	locs, err := s.geo.Query(coords)
	if err != nil {
		respondError(c, 50002, err.Error())
		return
	}
	respond(c, 0, "success", locs)
}

func runHTTPServer(geo *rgeocoder.RGeocoder, addr string) error {
	r := gin.Default()
	s := newAPIServer(geo)
	s.register(r)
	log.Printf("HTTP server listening on %s", addr)
	return r.Run(addr)
}
