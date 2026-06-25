package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	apiBaseURL = "https://api.bankless.com"
)

func getAPIKey() string {
	return os.Getenv("BANKLESS_API_TOKEN")
}

func makeAPIRequest(ctx context.Context, endpoint string, method string, body io.Reader) ([]byte, error) {
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, method, apiBaseURL+endpoint, body)
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BANKLESS-TOKEN", getAPIKey())

	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
}

	return io.ReadAll(resp.Body)
}

func HandleReadContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	contract, _ :=getString(args, "contract")
	method, _ :=getString(args, "method")

	var inputs []map[string]interface{}
	if rawInputs, found := args["inputs"]; found {
		if inputSlice, found := rawInputs.([]interface{}); found {
			for _, item := range inputSlice {
				if inputMap, found := item.(map[string]interface{}); found {
					inputs = append(inputs, inputMap)

			}
		}
	}

	var outputs []map[string]interface{}
	if rawOutputs, found := args["outputs"]; found {
		if outputSlice, found := rawOutputs.([]interface{}); found {
			for _, item := range outputSlice {
				if outputMap, found := item.(map[string]interface{}); found {
					outputs = append(outputs, outputMap)

			}
		}
	}

	payload := map[string]interface{}{
		"network":  network,
		"contract": contract,
		"method":   method,
		"inputs":   inputs,
		"outputs":  outputs,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/contracts/read", "POST", strings.NewReader(string(payloadBytes)))
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response []map[string]interface{}
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	return ok(fmt.Sprintf("%+v", response))
}

}
}

func HandleGetProxy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	contract, _ :=getString(args, "contract")

	params := url.Values{}
	params.Add("network", network)
	params.Add("contract", contract)

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/contracts/proxy?"+params.Encode(), "GET", nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response map[string]interface{}
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	implementation, found := response["implementation"].(string)
	if !found {
		return err("invalid response format")
}

	return ok(implementation)
}

func HandleGetEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	addresses, _ :=getString(args, "addresses")
	topic, _ :=getString(args, "topic")

	var optionalTopics []string
	if rawOptionalTopics, found := args["optionalTopics"]; found {
		if topicSlice, found := rawOptionalTopics.([]interface{}); found {
			for _, item := range topicSlice {
				if topicStr, found := item.(string); found {
					optionalTopics = append(optionalTopics, topicStr)

			}
		}
	}

	payload := map[string]interface{}{
		"network":        network,
		"addresses":      strings.Split(addresses, ","),
		"topic":          topic,
		"optionalTopics": optionalTopics,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/events", "POST", strings.NewReader(string(payloadBytes)))
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response map[string]interface{}
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	return ok(fmt.Sprintf("%+v", response))
}

}

func HandleBuildEventTopic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	name, _ :=getString(args, "name")

	var arguments []map[string]interface{}
	if rawArguments, found := args["arguments"]; found {
		if argumentSlice, found := rawArguments.([]interface{}); found {
			for _, item := range argumentSlice {
				if argumentMap, found := item.(map[string]interface{}); found {
					arguments = append(arguments, argumentMap)

			}
		}
	}

	payload := map[string]interface{}{
		"network":   network,
		"name":      name,
		"arguments": arguments,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/events/topic", "POST", strings.NewReader(string(payloadBytes)))
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response string
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	return ok(response)
}

}

func HandleGetTransactionHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	userAddress, _ :=getString(args, "userAddress")

	var contract string
	if rawContract, found := args["contract"]; found {
		contract = getString(args, "contract")

	var methodID string
	if rawMethodID, found := args["methodID"]; found {
		methodID = getString(args, "methodID")

	var startBlock int
	if rawStartBlock, found := args["startBlock"]; found {
		startBlock = getInt(args, "startBlock")

	includeData, _ :=getBool(args, "includeData")

	payload := map[string]interface{}{
		"network":      network,
		"userAddress":  userAddress,
		"contract":     contract,
		"methodID":     methodID,
		"startBlock":   startBlock,
		"includeData":  includeData,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/transactions/history", "POST", strings.NewReader(string(payloadBytes)))
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response []map[string]interface{}
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	return ok(fmt.Sprintf("%+v", response))
}

}
}
}

func HandleGetTransactionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	network, _ :=getString(args, "network")
	transactionHash, _ :=getString(args, "transactionHash")

	params := url.Values{}
	params.Add("network", network)
	params.Add("transactionHash", transactionHash)

	responseBytes, apiErr := makeAPIRequest(ctx, "/v1/transactions/info?"+params.Encode(), "GET", nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var response map[string]interface{}
	unmarshalErr := json.Unmarshal(responseBytes, &response)
	if unmarshalErr != nil {
		return err(unmarshalErr.Error())
}

	return ok(fmt.Sprintf("%+v", response))
}