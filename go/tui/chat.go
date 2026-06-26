package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NexusSoftMDMA/TormentNexus/agents"
)

type model struct {
	director *agents.Director
	input    string
	history  []string
	loading  bool
	spinner  spinner.Model
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		director: agents.NewDirector(agents.NewTormentNexusProvider()),
		input:    "",
		history:  []string{},
		loading:  false,
		spinner:  s,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if strings.TrimSpace(m.input) != "" {
				req := strings.TrimSpace(m.input)

				if strings.HasPrefix(m.input, "/") {
					m.input = ""
					mdl, cmd := ProcessSlashCommand(req, &m)
					m = mdl.(model)
					return m, cmd
				}

				if strings.HasPrefix(m.input, "??") {
					m.input = ""
					mdl, cmd := ProcessShellCommand(req, &m)
					m = mdl.(model)
					return m, cmd
				}

				m.history = append(m.history, "You: "+req)
				m.input = ""
				m.loading = true
				cmds = append(cmds, func() tea.Msg {
					response, err := buildPromptResponse(m.director, req)
					if err != nil {
						return fmt.Sprintf("Error: %v", err)
					}
					return response
				})
				return m, tea.Batch(cmds...)
			}
		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case tea.KeyRunes, tea.KeySpace:
			m.input += msg.String()
		}
	case string:
		m.loading = false
		m.history = append(m.history, "TormentNexus-Go-Director: "+msg)

	case PromptDisplayMsg:
		m.loading = false
		m.history = append(m.history, "TormentNexus-Go-Director: "+msg.Display)

	case ShellProposalMsg:
		m.loading = false
		display := fmt.Sprintf("[Shell Proposal] %s", msg.Command)
		if strings.TrimSpace(msg.Explanation) != "" {
			display += fmt.Sprintf("\n[Route] %s", msg.Explanation)
		}
		display += "\n> Execute? (Y/n)"
		m.history = append(m.history, display)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := strings.Join(m.history, "\n")
	s += "\n\n"
	if m.loading {
		s += fmt.Sprintf("%s Processing neural inputs...\n", m.spinner.View())
	} else {
		s += "> " + m.input
	}
	return s
}

func StartREPL() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
