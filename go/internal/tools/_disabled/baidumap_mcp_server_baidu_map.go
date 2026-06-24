package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ak, _ :=getString(args, "api_key")
	addr, _ :=getString(args, "address")
	u := fmt.Sprintf("https://api.map.baidu.com/geocoding/v3/?address=%s&output=json&ak=%s", url.QueryEscape(addr), ak)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var r map[string]interface{}
	if e := json.Unmarshal(body, &r); e != nil {
		return err("parse failed")
}

	return success(string(body))
}

func HandleReverseGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ak, _ :=getString(args, "api_key")
	lat, _ :=getString(args, "lat")
	lng, _ :=getString(args, "lng")
	u := fmt.Sprintf("https://api.map.baidu.com/reverse_geocoding/v3/?location=%s,%s&output=json&ak=%s", lat, lng, ak)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var r map[string]interface{}
	if e := json.Unmarshal(body, &r); e != nil {
		return err("parse failed")
}

	return success(string(body))
}