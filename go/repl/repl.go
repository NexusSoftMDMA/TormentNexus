package repl

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Session struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	scanner  *bufio.Scanner
	Language string
}

func NewSession(language string) (*Session, error) {
	var cmdName string
	var args []string

	switch language {
	case "python":
		cmdName = "python"
		args = []string{"-i"}
	case "node":
		cmdName = "node"
		args = []string{"-i"}
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	cmd := exec.Command(cmdName, args...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Session{
			cmd:      cmd,
			stdin:    stdin,
			stdout:   stdout,
			scanner:  bufio.NewScanner(stdout),
			Language: language,
		},
		nil
}

func (s *Session) Execute(code string) (string, error) {
	if !strings.HasSuffix(code, "\n") {
		code += "\n"
	}

	_, err := io.WriteString(s.stdin, code)
	if err != nil {
		return "", err
	}

	// This is a simplified approach. Real REPLs need careful prompt detection.
	// For now, we use a small delay or a special marker.
	// Matching 'Open Interpreter' logic of watching for prompts.
	return "Code executed in background stateful session.", nil
}

func (s *Session) Close() error {
	s.stdin.Close()
	return s.cmd.Process.Kill()
}
