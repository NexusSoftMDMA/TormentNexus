package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

const earthRadiusKm = 6371.0

func HandleCalculateDistance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat1S, _ :=getString(args, "lat1")
	lon1S, _ :=getString(args, "lon1")
	lat2S, _ :=getString(args, "lat2")
	lon2S, _ :=getString(args, "lon2")
	if lat1S == "" || lon1S == "" || lat2S == "" || lon2S == "" {
		return err("missing coordinate")
}

	lat1, e := strconv.ParseFloat(lat1S, 64)
	if e != nil {
		return err("invalid lat1")
}

	lon1, e := strconv.ParseFloat(lon1S, 64)
	if e != nil {
		return err("invalid lon1")
}

	lat2, e := strconv.ParseFloat(lat2S, 64)
	if e != nil {
		return err("invalid lat2")
}

	lon2, e := strconv.ParseFloat(lon2S, 64)
	if e != nil {
		return err("invalid lon2")
}

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	latMid := (lat1Rad + lat2Rad) / 2
	distance := earthRadiusKm * math.Sqrt(dLat*dLat+math.Cos(latMid)*math.Cos(latMid)*dLon*dLon)
	result, _ := json.Marshal(map[string]float64{"distance_km": math.Round(distance*100) / 100})
	return success(string(result))
}

func HandleCalculateBearing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat1S, _ :=getString(args, "lat1")
	lon1S, _ :=getString(args, "lon1")
	lat2S, _ :=getString(args, "lat2")
	lon2S, _ :=getString(args, "lon2")
	if lat1S == "" || lon1S == "" || lat2S == "" || lon2S == "" {
		return err("missing coordinate")
}

	lat1, e := strconv.ParseFloat(lat1S, 64)
	if e != nil {
		return err("invalid lat1")
}

	lon1, e := strconv.ParseFloat(lon1S, 64)
	if e != nil {
		return err("invalid lon1")
}

	lat2, e := strconv.ParseFloat(lat2S, 64)
	if e != nil {
		return err("invalid lat2")
}

	lon2, e := strconv.ParseFloat(lon2S, 64)
	if e != nil {
		return err("invalid lon2")
}

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	y := math.Sin(dLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLon)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}
	result, _ := json.Marshal(map[string]float64{"bearing_deg": math.Round(bearing*100) / 100})
	return success(string(result))
}