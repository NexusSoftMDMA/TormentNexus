package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"regexp"
	"sort"
)

func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if e := validateArgs(args); e != nil {
		return getToolResponseError(e.Error())
}

	return getToolResponseOK("Success")
}

func HandleYyy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if e := validateArgs(args); e != nil {
		return getToolResponseError(e.Error())
}

	return getToolResponseOK("Success")
}

func HandleZzz(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if e := validateArgs(args); e != nil {
		return getToolResponseError(e.Error())
}

	return getToolResponseOK("Success")
}

func HandleAaa(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if e := validateArgs(args); e != nil {
		return getToolResponseError(e.Error())
}

	return getToolResponseOK("Success")
}

func validateArgs(args map[string]interface{}) error {
	requiredKeys := []string{"key1", "key2"}
	for _, key := range requiredKeys {
		_, exists := args[key]
		if !exists {
			return fmt.Errorf("missing required argument: %s", key)

	}
	return nil
}

}

func getToolResponseOK(text string) (ToolResponse, error) {
	return ok(text)
}

func getToolResponseError(text string) (ToolResponse, error) {
	return err(text)
}

func getString(args map[string]interface{}, key string) string {
	val, found := args[key]
	if !found {
		return ""
	}
	strVal, found := val.(string)
	if !found {
		return ""
	}
	return strVal
}

func getInt(args map[string]interface{}, key string) int {
	strVal, _ :=getString(args, key)
	intVal, _ := strconv.Atoi(strVal)
	return intVal
}

func getBool(args map[string]interface{}, key string) bool {
	strVal, _ :=getString(args, key)
	boolVal, _ := strconv.ParseBool(strVal)
	return boolVal
}