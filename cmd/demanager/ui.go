package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// messages
type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// UIModel for BubbleTea
type UIModel struct {
	AppState      *AppState
	CurrentTabIdx int
	Width         int
	Height        int
	FocusedButton int // 0 = Stop, 1 = Start, 2 = Restart All
}

func NewUIModel(appState *AppState) UIModel {
	return UIModel{
		AppState:      appState,
		CurrentTabIdx: 0,
		FocusedButton: 0,
	}
}

func (m UIModel) Init() tea.Cmd {
	return tickCmd()
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left":
			if m.CurrentTabIdx > 0 {
				m.CurrentTabIdx--
			}

		case "right":
			if m.CurrentTabIdx < len(m.AppState.OrderedNames)-1 {
				m.CurrentTabIdx++
			}

		case "tab":
			m.FocusedButton = (m.FocusedButton + 1) % 3

		case "shift+tab":
			m.FocusedButton = (m.FocusedButton - 1 + 3) % 3

		case "enter", " ":
			currentService := m.AppState.OrderedNames[m.CurrentTabIdx]
			switch m.FocusedButton {
			case 0:
				StopService(m.AppState, currentService)
			case 1:
				rs := m.AppState.RunningServices[currentService]
				go func() {
					err := rs.Start()
					if err != nil {
						rs.AddLog(fmt.Sprintf("[ERROR] Failed to start: %v\n", err))
					}
				}()
			case 2:
				go func() {
					err := RestartAllServices(m.AppState)
					if err != nil {
						fmt.Println("Error restarting all:", err)
					}
				}()
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case TickMsg:
		return m, tickCmd()
	}

	return m, nil
}

func (m UIModel) View() string {
	if m.Width == 0 {
		return "Loading..."
	}

	tabsView := m.renderTabs()

	logsView := m.renderLogs()

	buttonsView := m.renderButtons()

	return fmt.Sprintf("%s\n\n%s\n\n%s", tabsView, logsView, buttonsView)
}

func (m UIModel) renderTabs() string {
	var tabs []string
	for i, sName := range m.AppState.OrderedNames {
		if i == m.CurrentTabIdx {
			tabs = append(tabs, fmt.Sprintf("[%s]", sName))
		} else {
			tabs = append(tabs, fmt.Sprintf(" %s ", sName))
		}
	}

	tabsStr := strings.Join(tabs, " ")
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Render(tabsStr)
}

func (m UIModel) renderLogs() string {
	if len(m.AppState.OrderedNames) == 0 {
		return "No services"
	}

	currentService := m.AppState.OrderedNames[m.CurrentTabIdx]
	rs := m.AppState.RunningServices[currentService]

	logs := rs.GetLogs()

	// show last 30 logs that fit in the available height
	maxHeight := 30
	startIdx := len(logs) - maxHeight
	if startIdx < 0 {
		startIdx = 0
	}

	visibleLogs := logs[startIdx:]
	logsStr := strings.Join(visibleLogs, "")

	// truncate lines to max width of 120 chars
	maxWidth := 120
	lines := strings.Split(logsStr, "\n")
	var truncatedLines []string
	for _, line := range lines {
		if len(line) > maxWidth {
			truncatedLines = append(truncatedLines, line[:maxWidth])
		} else {
			truncatedLines = append(truncatedLines, line)
		}
	}
	logsStr = strings.Join(truncatedLines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(120).
		Height(30).
		Render(logsStr)
}

func (m UIModel) renderButtons() string {
	buttons := []string{
		m.renderButton(0, "Stop"),
		m.renderButton(1, "Start"),
		m.renderButton(2, "Restart All"),
	}

	return strings.Join(buttons, "\n")
}

func (m UIModel) renderButton(idx int, label string) string {
	style := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())

	if idx == m.FocusedButton {
		style = style.
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("0"))
	} else {
		style = style.
			Foreground(lipgloss.Color("4"))
	}

	return style.Render(label)
}

func RunUI(appState *AppState) error {
	model := NewUIModel(appState)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
