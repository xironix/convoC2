package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cxnturi0n/convoC2/pkg/server"
)

func (m model) Init() tea.Cmd {
	return WaitForAgent(m.agentChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	type commandOutput string

	switch msg := msg.(type) {
	case server.Agent:

		if contains(m.agents, msg) {
			return m, WaitForAgent(m.agentChan)
		}
		m.agents = append(m.agents, msg)
		m.list.InsertItem(len(m.agents)-1, item(fmt.Sprintf("%s (%s)", msg.Username, msg.AgentId)))
		return m, WaitForAgent(m.agentChan)

	case commandOutput:
		if m.screen == AgentCmdScreen {
			m.selectedAgent.CommandHistoryCmd = append(m.selectedAgent.CommandHistoryCmd, string(msg))
		} else {
			m.selectedAgent.CommandHistory = append(m.selectedAgent.CommandHistory, string(msg))
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.screen == AgentListScreen {
				agentsNum := len(m.list.Items())
				if agentsNum > 0 {
					selectedAgentIndex := m.list.Index()
					m.selectedAgent = &m.agents[selectedAgentIndex]
					m.screen = AgentScreen
				}
			} else if m.screen == AgentScreen {
				enteredCommand := m.textInput.Value()
				m.handleRegularCommands(enteredCommand)
				m.textInput.Reset()
			} else if m.screen == AgentCmdScreen {
				enteredCommand := m.textInput.Value()
				m.handleCmdSession(enteredCommand)
				m.textInput.Reset()
			}
		}

		if m.screen == AgentScreen || m.screen == AgentCmdScreen {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	if m.screen == AgentListScreen {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.screen == AgentListScreen {
		return bannerStyle.Render(banner) + "\n" + m.list.View()
	} else if m.screen == AgentScreen || m.screen == AgentCmdScreen {
		historyView := ""
		if m.screen == AgentCmdScreen {
			historyView = strings.Join(m.selectedAgent.CommandHistoryCmd, "\n")
		} else {
			historyView = strings.Join(m.selectedAgent.CommandHistory, "\n")
		}

		if m.screen != AgentCmdScreen {
			agentDetails := agentHeaderStyle.Render(fmt.Sprintf("Agent: %s - (%s)", m.selectedAgent.Username, m.selectedAgent.AgentId))
			if historyView == "" {
				return fmt.Sprintf("%s\n\n%s", agentDetails, m.textInput.View())
			}
			return fmt.Sprintf("%s\n\n%s\n%s", agentDetails, historyView, m.textInput.View())
		}

		if historyView == "" {
			return m.textInput.View()
		}
		return fmt.Sprintf("%s\n%s", historyView, m.textInput.View())
	}

	return ""
}

func contains(agents []server.Agent, agent server.Agent) bool {
	for _, item := range agents {
		if item.AgentId == agent.AgentId {
			return true
		}
	}
	return false
}
