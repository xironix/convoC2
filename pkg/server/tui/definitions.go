package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cxnturi0n/convoC2/pkg/server"
)

const banner = `
░▒▓█▓▒░░▒▓█▓▒░░▒▓████████▓▒░░▒▓███████▓▒░  ░▒▓██████▓▒░░▒▓████████▓▒░░▒▓█▓▒░ ░▒▓██████▓▒░ ░▒▓███████▓▒░        ░▒▓██████▓▒░ ░▒▓███████▓▒░  
░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░       ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░  ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░       ░▒▓█▓▒░ 
░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░       ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░  ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░              ░▒▓█▓▒░ 
 ░▒▓██████▓▒░ ░▒▓██████▓▒░  ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░  ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓████████▓▒░░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░        ░▒▓██████▓▒░  
░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░       ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░  ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░       ░▒▓█▓▒░        
░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░       ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░  ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░        
░▒▓█▓▒░░▒▓█▓▒░░▒▓████████▓▒░░▒▓█▓▒░░▒▓█▓▒░ ░▒▓██████▓▒░   ░▒▓█▓▒░    ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░       ░▒▓██████▓▒░ ░▒▓████████▓▒░ 
                                                                                                                                           
By XenoFloozy`

const listHeight = 14
const listDefaultWidth = 300

const (
	AgentListScreen = 0
	AgentScreen     = 1
	AgentCmdScreen  = 2
)

var (
	titleStyle        = lipgloss.NewStyle()
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF1C1C"))
	bannerStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4c00a4"))
	agentHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4c00a4")).Bold(true)
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#20C20E"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1C1C"))
	agentPromptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1C1C"))
	promptStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1C1C"))
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

type model struct {
	screen              int
	agents              []server.Agent
	list                list.Model
	agentChan           chan server.Agent
	selectedAgent       *server.Agent
	textInput           textinput.Model
	showHelp            bool
	commandResponseChan chan server.CommandResponse
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(str))
}

func WaitForAgent(agentChan chan server.Agent) tea.Cmd {
	return func() tea.Msg {
		return <-agentChan
	}
}

func InitialModel() model {
	items := []list.Item{}
	l := list.New(items, itemDelegate{}, listDefaultWidth, listHeight)
	l.Title = "\nAgents connected:"

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetStatusBarItemName("agent", "agents")

	ti := textinput.New()
	ti.Placeholder = "Type ? for a list of commands"
	ti.Prompt = promptStyle.Render(">> ")
	ti.Focus()
	ti.Width = 50

	m := model{
		screen:              AgentListScreen,
		list:                l,
		textInput:           ti,
		agentChan:           make(chan server.Agent),
		agents:              []server.Agent{},
		commandResponseChan: make(chan server.CommandResponse),
		showHelp:            false,
	}

	go WaitForAgent(m.agentChan)
	go server.StartHttpListener(m.agentChan, m.commandResponseChan)

	return m
}
