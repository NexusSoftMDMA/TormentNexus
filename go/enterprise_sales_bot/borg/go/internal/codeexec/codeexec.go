package codeexec

type Sandbox struct {
	path string
}

func NewSandbox(path string) *Sandbox {
	return &Sandbox{path: path}
}

type ExecResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (s *Sandbox) Execute(language, code string) (*ExecResult, error) {
	// Stub implementation
	return &ExecResult{Output: "Stub executed"}, nil
}
