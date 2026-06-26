package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type JsonRpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JsonRpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type StdioClient struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	pendingMu sync.Mutex
	pending   map[interface{}]chan *JsonRpcResponse
	nextID    int
}

func NewStdioClient(name, command string, args []string, env map[string]string) *StdioClient {
	return &StdioClient{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
		pending: make(map[interface{}]chan *JsonRpcResponse),
	}
}

func (c *StdioClient) Start() error {
	c.cmd = exec.Command(c.Command, c.Args...)

	// Setup env by extending the current process environment so child processes
	// keep required platform/runtime variables (important on Windows).
	c.cmd.Env = append([]string{}, os.Environ()...)
	for k, v := range c.Env {
		c.cmd.Env = append(c.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return err
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	c.stdout = stdout

	if err := c.cmd.Start(); err != nil {
		return err
	}

	go c.readLoop()

	return nil
}

func (c *StdioClient) readLoop() {
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		var resp JsonRpcResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err == nil {
			c.pendingMu.Lock()
			if ch, ok := c.pending[resp.ID]; ok {
				ch <- &resp
				delete(c.pending, resp.ID)
			}
			c.pendingMu.Unlock()
		}
	}
}

func (c *StdioClient) Call(ctx context.Context, method string, params interface{}) (*JsonRpcResponse, error) {
	c.pendingMu.Lock()
	c.nextID++
	id := fmt.Sprintf("%d", c.nextID)
	ch := make(chan *JsonRpcResponse, 1)
	c.pending[id] = ch
	c.pendingMu.Unlock()

	req := JsonRpcRequest{
		Jsonrpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, _ := json.Marshal(req)
	fmt.Fprintln(c.stdin, string(data))

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *StdioClient) Stop() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}
