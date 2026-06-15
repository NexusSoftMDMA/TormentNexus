//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var nwsBaseURL = func() string {
	if url := os.Getenv("NWS_API_URL"); url != "" {
		return url
	}
	return "https://api.weather.gov"
}()

// Points metadata cache
type pointsCacheEntry struct {
	data    map[string]interface{}
	expires time.Time
}

var (
	pointsCache   = make(map[string]pointsCacheEntry)
	pointsCacheMu sync.RWMutex
)

// Helper: Haversine distance in kilometers
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180.0 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Helper: Bearing in degrees
func bearing(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(deg float64) float64 { return deg * math.Pi / 180.0 }
	dLon := toRad(lon2 - lon1)
	y := math.Sin(dLon) * math.Cos(toRad(lat2))
	x := math.Cos(toRad(lat1))*math.Sin(toRad(lat2)) -
		math.Sin(toRad(lat1))*math.Cos(toRad(lat2))*math.Cos(dLon)
	brng := math.Atan2(y, x) * 180.0 / math.Pi
	return math.Mod(brng+360.0, 360.0)
}

// Helper: Bearing degrees to compass direction
func bearingToCompass(deg float64) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int(math.Round(deg/22.5)) % 16
	return dirs[idx]
}

// callNWSAPI performs HTTP GET requests to api.weather.gov
func callNWSAPI(ctx context.Context, urlPath string, queryParams map[string]string) ([]byte, error) {
	reqUrl := urlPath
	if !strings.HasPrefix(reqUrl, "http") {
		reqUrl = nwsBaseURL + urlPath
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "(nws-weather-mcp-server, contact@cyanheads.com)")
	req.Header.Set("Accept", "application/geo+json")

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			if v != "" {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NWS API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Resolve points coordinates to metadata
func resolveNWSPoints(ctx context.Context, lat, lon float64) (map[string]interface{}, error) {
	// Round coordinates to 4 decimal places per NWS API recommendation
	tLat := math.Round(lat*10000) / 10000
	tLon := math.Round(lon*10000) / 10000
	cacheKey := fmt.Sprintf("%.4f_%.4f", tLat, tLon)

	pointsCacheMu.RLock()
	entry, found := pointsCache[cacheKey]
	pointsCacheMu.RUnlock()

	if found && entry.expires.After(time.Now()) {
		return entry.data, nil
	}

	path := fmt.Sprintf("/points/%.4f,%.4f", tLat, tLon)
	respData, err := callNWSAPI(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(respData, &parsed); err != nil {
		return nil, err
	}

	properties, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("NWS points response missing properties")
	}

	pointsCacheMu.Lock()
	pointsCache[cacheKey] = pointsCacheEntry{
		data:    properties,
		expires: time.Now().Add(1 * time.Hour), // Cache for 1 hour
	}
	pointsCacheMu.Unlock()

	return properties, nil
}

// Helper: extract ID or zone code from full URL path segment
func extractLastSegment(urlStr interface{}) string {
	if s, ok := urlStr.(string); ok {
		parts := strings.Split(strings.TrimRight(s, "/"), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// HandleNWSGetForecast retrieves weather forecast for coordinates
func HandleNWSGetForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat := getFloat(args, "latitude")
	lon := getFloat(args, "longitude")
	hourly := getBool(args, "hourly")

	if lat == 0.0 && lon == 0.0 {
		return err("latitude and longitude parameters are required and must be non-zero")
	}

	points, errVal := resolveNWSPoints(ctx, lat, lon)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	forecastURLKey := "forecast"
	if hourly {
		forecastURLKey = "forecastHourly"
	}

	forecastURL, okForecast := points[forecastURLKey].(string)
	if !okForecast || forecastURL == "" {
		return err(fmt.Sprintf("NWS points response missing forecast URL: %s", forecastURLKey))
	}

	respData, errVal := callNWSAPI(ctx, forecastURL, nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	// Try parsing and packaging output formatted closely to cyanheads tool
	var rawForecast map[string]interface{}
	if err := json.Unmarshal(respData, &rawForecast); err == nil {
		properties, _ := rawForecast["properties"].(map[string]interface{})
		relativeLocation, _ := points["relativeLocation"].(map[string]interface{})
		relProps, _ := relativeLocation["properties"].(map[string]interface{})

		city, _ := relProps["city"].(string)
		state, _ := relProps["state"].(string)
		office, _ := points["gridId"].(string)
		timeZone, _ := points["timeZone"].(string)
		forecastZone := extractLastSegment(points["forecastZone"])
		county := extractLastSegment(points["county"])

		output := map[string]interface{}{
			"location": map[string]interface{}{
				"city":         city,
				"state":        state,
				"office":       office,
				"timeZone":     timeZone,
				"forecastZone": forecastZone,
				"county":       county,
			},
			"generatedAt": properties["generatedAt"],
			"periods":     properties["periods"],
		}
		b, errMarshal := json.MarshalIndent(output, "", "  ")
		if errMarshal == nil {
			return ok(string(b))
		}
	}

	return ok(string(respData))
}

// HandleNWSSearchAlerts searches active weather alerts
func HandleNWSSearchAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queryParams := make(map[string]string)

	if area, _ := getString(args, "area"); area != "" {
		queryParams["area"] = area
	}
	if point, _ := getString(args, "point"); point != "" {
		queryParams["point"] = point
	}
	if zone, _ := getString(args, "zone"); zone != "" {
		queryParams["zone"] = zone
	}
	if status, _ := getString(args, "status"); status != "" {
		queryParams["status"] = status
	}

	// Array/slice parameters
	if severity := getStringSlice(args, "severity"); len(severity) > 0 {
		queryParams["severity"] = strings.Join(severity, ",")
	}
	if urgency := getStringSlice(args, "urgency"); len(urgency) > 0 {
		queryParams["urgency"] = strings.Join(urgency, ",")
	}
	if certainty := getStringSlice(args, "certainty"); len(certainty) > 0 {
		queryParams["certainty"] = strings.Join(certainty, ",")
	}

	respData, errVal := callNWSAPI(ctx, "/alerts/active", queryParams)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	// Apply filter by event client side if required
	eventFilters := getStringSlice(args, "event")
	if len(eventFilters) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(respData, &parsed); err == nil {
			features, _ := parsed["features"].([]interface{})
			var filtered []interface{}
			for _, f := range features {
				if fMap, ok := f.(map[string]interface{}); ok {
					props, _ := fMap["properties"].(map[string]interface{})
					evt, _ := props["event"].(string)
					evtLower := strings.ToLower(evt)
					match := false
					for _, filter := range eventFilters {
						if strings.Contains(evtLower, strings.ToLower(filter)) {
							match = true
							break
						}
					}
					if match {
						filtered = append(filtered, f)
					}
				}
			}
			parsed["features"] = filtered
			b, _ := json.MarshalIndent(parsed, "", "  ")
			return ok(string(b))
		}
	}

	return ok(string(respData))
}

// HandleNWSGetObservations retrieves current measured conditions
func HandleNWSGetObservations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	stationId, _ := getString(args, "station_id", "stationId")
	lat := getFloat(args, "latitude")
	lon := getFloat(args, "longitude")

	if stationId == "" && (lat == 0.0 && lon == 0.0) {
		return err("either station_id or latitude/longitude coordinates must be provided")
	}

	if stationId == "" {
		// Sequentially resolve nearest station ID from coordinates
		points, errVal := resolveNWSPoints(ctx, lat, lon)
		if errVal != nil {
			return errResponseNWS(errVal)
		}
		stationsURL, okStations := points["observationStations"].(string)
		if !okStations || stationsURL == "" {
			return err("NWS points response missing observationStations URL")
		}

		stationsResp, errVal := callNWSAPI(ctx, stationsURL, nil)
		if errVal != nil {
			return errResponseNWS(errVal)
		}

		var parsedStations map[string]interface{}
		if err := json.Unmarshal(stationsResp, &parsedStations); err != nil {
			return errResponseNWS(err)
		}

		features, _ := parsedStations["features"].([]interface{})
		if len(features) == 0 {
			return err("no observation stations found near this location")
		}

		// Pick nearest station
		var nearestStationId string
		minDist := math.MaxFloat64
		for _, f := range features {
			fMap, _ := f.(map[string]interface{})
			props, _ := fMap["properties"].(map[string]interface{})
			sId, _ := props["stationIdentifier"].(string)
			geom, _ := fMap["geometry"].(map[string]interface{})
			coords, _ := geom["coordinates"].([]interface{})
			if len(coords) >= 2 {
				sLon, _ := coords[0].(float64)
				sLat, _ := coords[1].(float64)
				dist := haversine(lat, lon, sLat, sLon)
				if dist < minDist {
					minDist = dist
					nearestStationId = sId
				}
			}
		}

		if nearestStationId == "" {
			return err("unable to calculate nearest observation station")
		}
		stationId = nearestStationId
	}

	path := fmt.Sprintf("/stations/%s/observations/latest", strings.ToUpper(stationId))
	respData, errVal := callNWSAPI(ctx, path, nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	return ok(string(respData))
}

// HandleNWSFindStations discovers nearby observation stations
func HandleNWSFindStations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat := getFloat(args, "latitude")
	lon := getFloat(args, "longitude")
	limit := getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	if lat == 0.0 && lon == 0.0 {
		return err("latitude and longitude parameters are required and must be non-zero")
	}

	points, errVal := resolveNWSPoints(ctx, lat, lon)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	stationsURL, okStations := points["observationStations"].(string)
	if !okStations || stationsURL == "" {
		return err("NWS points response missing observationStations URL")
	}

	respData, errVal := callNWSAPI(ctx, stationsURL, nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(respData, &parsed); err == nil {
		features, _ := parsed["features"].([]interface{})
		type StationDistance struct {
			Feature  interface{}
			Distance float64
			Bearing  float64
		}
		var list []StationDistance
		for _, f := range features {
			fMap, _ := f.(map[string]interface{})
			geom, _ := fMap["geometry"].(map[string]interface{})
			coords, _ := geom["coordinates"].([]interface{})
			if len(coords) >= 2 {
				sLon, _ := coords[0].(float64)
				sLat, _ := coords[1].(float64)
				dist := haversine(lat, lon, sLat, sLon)
				brng := bearing(lat, lon, sLat, sLon)
				list = append(list, StationDistance{Feature: f, Distance: dist, Bearing: brng})
			}
		}

		sort.Slice(list, func(i, j int) bool {
			return list[i].Distance < list[j].Distance
		})

		var result []interface{}
		for i := 0; i < len(list) && i < limit; i++ {
			fMap, _ := list[i].Feature.(map[string]interface{})
			props, _ := fMap["properties"].(map[string]interface{})
			sId, _ := props["stationIdentifier"].(string)
			name, _ := props["name"].(string)
			elevation, _ := props["elevation"].(map[string]interface{})
			timeZone, _ := props["timeZone"].(string)
			county := extractLastSegment(props["county"])
			forecastZone := extractLastSegment(props["forecast"])

			result = append(result, map[string]interface{}{
				"stationId":    sId,
				"name":         name,
				"distance":     math.Round(list[i].Distance*10) / 10,
				"bearing":      bearingToCompass(list[i].Bearing),
				"elevation":    elevation,
				"timeZone":     timeZone,
				"county":       county,
				"forecastZone": forecastZone,
			})
		}

		output := map[string]interface{}{
			"stations":   result,
			"totalFound": len(list),
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		return ok(string(b))
	}

	return ok(string(respData))
}

// HandleNWSListAlertTypes lists valid alert event type names
func HandleNWSListAlertTypes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	respData, errVal := callNWSAPI(ctx, "/alerts/types", nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}
	return ok(string(respData))
}

// HandleNWSGetOfficeDiscussion fetches weather discussions from WFO
func HandleNWSGetOfficeDiscussion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	office, _ := getString(args, "office")
	productType, _ := getString(args, "product_type", "productType")
	if productType == "" {
		productType = "AFD"
	}

	if office == "" {
		return err("office parameter is required")
	}

	office = strings.ToUpper(strings.TrimSpace(office))
	productType = strings.ToUpper(strings.TrimSpace(productType))

	// Step 1: list products
	listPath := fmt.Sprintf("/products/types/%s/locations/%s", productType, office)
	listData, errVal := callNWSAPI(ctx, listPath, nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	var listParsed map[string]interface{}
	if err := json.Unmarshal(listData, &listParsed); err != nil {
		return errResponseNWS(err)
	}

	graph, _ := listParsed["@graph"].([]interface{})
	if len(graph) == 0 {
		return err(fmt.Sprintf("No %s products found for office %s.", productType, office))
	}

	// Pick latest product ID
	latest, _ := graph[0].(map[string]interface{})
	id, _ := latest["id"].(string)
	if id == "" {
		return err("missing ID in latest product list entry")
	}

	// Step 2: fetch detail
	detailData, errVal := callNWSAPI(ctx, fmt.Sprintf("/products/%s", id), nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	var detailParsed map[string]interface{}
	if err := json.Unmarshal(detailData, &detailParsed); err == nil {
		output := map[string]interface{}{
			"issuanceTime":    detailParsed["issuanceTime"],
			"issuingOffice":   detailParsed["issuingOffice"],
			"productCode":     detailParsed["productCode"],
			"productName":     detailParsed["productName"],
			"productText":     detailParsed["productText"],
			"wmoCollectiveId": detailParsed["wmoCollectiveId"],
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		return ok(string(b))
	}

	return ok(string(detailData))
}

// HandleNWSGetZoneForecast gets text forecast periods for a public zone
func HandleNWSGetZoneForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zoneId, _ := getString(args, "zone_id", "zoneId")
	if zoneId == "" {
		return err("zone_id parameter is required")
	}
	zoneId = strings.ToUpper(strings.TrimSpace(zoneId))

	path := fmt.Sprintf("/zones/forecast/%s/forecast", zoneId)
	respData, errVal := callNWSAPI(ctx, path, nil)
	if errVal != nil {
		return errResponseNWS(errVal)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(respData, &parsed); err == nil {
		properties, _ := parsed["properties"].(map[string]interface{})
		rawPeriods, _ := properties["periods"].([]interface{})
		var periods []interface{}
		for _, p := range rawPeriods {
			if pMap, ok := p.(map[string]interface{}); ok {
				periods = append(periods, map[string]interface{}{
					"number":           pMap["number"],
					"name":             pMap["name"],
					"detailedForecast": pMap["detailedForecast"],
				})
			}
		}
		output := map[string]interface{}{
			"zoneId":  zoneId,
			"updated": properties["updated"],
			"periods": periods,
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		return ok(string(b))
	}

	return ok(string(respData))
}

func getFloat(args map[string]interface{}, key string) float64 {
	if val, ok := args[key]; ok {
		if f, okF := val.(float64); okF {
			return f
		}
	}
	return 0.0
}

func errResponseNWS(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf("NWS API failed: %v", err)}},
		IsError: true,
	}, nil
}
