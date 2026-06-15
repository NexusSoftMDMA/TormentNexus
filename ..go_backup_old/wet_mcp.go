package tools

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "strings"
)

func HandleWetness(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    location, _ :=getString(args, "location")
    if location == "" {
        return err("location is required")
}

    url := fmt.Sprintf("https://wttr.in/%s?format=%%C", location)
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch weather: " + e.Error())
}

    defer resp.Body.Close()
    body, e := io.ReadAll(resp.Body)
    if e != nil {
        return err("failed to read response: " + e.Error())
}

    condition := strings.ToLower(strings.TrimSpace(string(body)))
    if strings.Contains(condition, "rain") || strings.Contains(condition, "drizzle") || strings.Contains(condition, "thunderstorm") {
        return ok("It is wet in " + location)
}

    return ok("It is not wet in " + location)
}

func HandleWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    city, _ :=getString(args, "city")
    if city == "" {
        return err("city is required")
}

    return ok("The weather in " + city + " is sunny")
}