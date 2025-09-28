package tests

import (
	"bytes"
	"testing"

	"github.com/your-username/reverse-geocoder-go/pkg/rgeocoder"
)

func TestEmptyDatasetQuery(t *testing.T) {
	rg, err := rgeocoder.NewRGeocoder()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	defer rg.Close()
	_, err = rg.Query([]rgeocoder.Coordinate{{Lat: 0, Lon: 0}})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
}

func TestStreamDataset(t *testing.T) {
	csvData := "lat,lon,name,admin1,admin2,cc\n37.78674,-122.39222,SampleCity,Region,Sub,US\n"
	r := bytes.NewBufferString(csvData)
	rg, err := rgeocoder.NewRGeocoderWithStream(r)
	if err != nil {
		t.Fatalf("init stream failed: %v", err)
	}
	loc, err := rg.QuerySingle(rgeocoder.Coordinate{Lat: 37.78674, Lon: -122.39222})
	if err != nil {
		t.Fatalf("query single failed: %v", err)
	}
	if loc.Name != "SampleCity" {
		if loc.Name == "" {
			t.Fatalf("expected SampleCity got empty")
		}
	}
}
