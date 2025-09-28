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

func (s *apiServer) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
}

// /reverse?lat=..&lon=..
func (s *apiServer) reverse(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing lat or lon"})
		return
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat or lon"})
		return
	}
	loc, err := s.geo.QuerySingle(rgeocoder.Coordinate{Lat: lat, Lon: lon})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loc)
}

type batchRequest struct {
	Points []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"points"`
}

type batchResponse struct {
	Results []rgeocoder.Location `json:"results"`
}

func (s *apiServer) batch(c *gin.Context) {
	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	coords := make([]rgeocoder.Coordinate, 0, len(req.Points))
	for _, p := range req.Points {
		coords = append(coords, rgeocoder.Coordinate{Lat: p.Lat, Lon: p.Lon})
	}
	locs, err := s.geo.Query(coords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batchResponse{Results: locs})
}

func runHTTPServer(geo *rgeocoder.RGeocoder, addr string) error {
	r := gin.Default()
	s := newAPIServer(geo)
	s.register(r)
	log.Printf("HTTP server listening on %s", addr)
	return r.Run(addr)
}
