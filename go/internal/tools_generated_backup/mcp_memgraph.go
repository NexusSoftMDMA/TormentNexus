package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type memgraphReq struct {
	Query  string                 `json:"query"`
	Params map[string]interface{} `json:"params"`
}

type memgraphResp struct {
	Results []map[string]interface{} `json:"results"`
	Error   string                   `json:"error,omitempty"`
}