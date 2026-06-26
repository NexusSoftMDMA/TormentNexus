package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ProcessShellCommand mimics Copilot CLI's `??` bash translation interceptor.
func ProcessShellCommand(cmd string, m *model) (tea.Model, tea.Cmd) {
	query := strings.TrimSpace(strings.TrimPrefix(cmd, "??"))

	m.history = append(m.history, fmt.Sprintf("You: ?? %s", query))

	m.loading = true
	var cmds []tea.Cmd

	cmds = append(cmds, func() tea.Msg {
		response, err := buildShellProposal(m.director, query)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return response
	})

	return *m, tea.Batch(cmds...)
}

// ShellProposalMsg represents a generated shell string waiting for user approval
type ShellProposalMsg struct {
	Command     string
	Explanation string
}
