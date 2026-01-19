package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kesavan-vaisakh/cmdfy/pkg/model"
)

// Style definitions
var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle  = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	colStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 1).
			Width(40)

	selectedColStyle = colStyle.
				BorderForeground(lipgloss.Color("#ADFF2F")). // Green-ish for selected
				Border(lipgloss.DoubleBorder())

	cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")). // Green code
			Bold(true)

	dangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

type ProviderResult struct {
	Name   string
	Result *model.CommandResult
	Error  error
}

type Model struct {
	Results  []ProviderResult
	Selected int
	Width    int
	Height   int
	Quitting bool
	Choice   *ProviderResult
}

func InitialModel(results []ProviderResult) Model {
	return Model{
		Results:  results,
		Selected: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "right", "l", "tab":
			if m.Selected < len(m.Results)-1 {
				m.Selected++
			} else {
				m.Selected = 0 // Wrap around
			}
		case "left", "h", "shift+tab":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Results) - 1 // Wrap around
			}
		case "enter":
			if len(m.Results) > 0 {
				m.Choice = &m.Results[m.Selected]
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	var columns []string

	for i, res := range m.Results {
		style := colStyle
		if i == m.Selected {
			style = selectedColStyle
		}

		title := titleStyle.Render(strings.ToUpper(res.Name))

		var content string
		if res.Error != nil {
			content = fmt.Sprintf("Error:\n%s", res.Error.Error())
		} else {
			cmdStr := ""
			for _, s := range res.Result.Steps {
				cmdStr += fmt.Sprintf("%s %s %s ", s.Tool, strings.Join(s.Args, " "), s.Op)
			}

			metrics := fmt.Sprintf("Latency: %s\nTokens: %d", res.Result.Metrics.Latency, res.Result.Metrics.TokenCount)

			explanation := lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render(res.Result.Explanation)

			dangerous := ""
			if res.Result.Dangerous {
				dangerous = dangerStyle.Render("\n[DANGEROUS]")
			}

			content = fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
				cmdStyle.Render(cmdStr),
				explanation,
				subtleStyle.Render(metrics),
				dangerous,
			)
		}

		col := style.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				title,
				content,
			),
		)
		columns = append(columns, col)
	}

	// Join columns horizontally
	ui := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Add help text
	help := subtleStyle.Render("\nUse arrow keys to navigate • Enter to select • q to quit")

	return lipgloss.JoinVertical(lipgloss.Center, ui, help)
}
