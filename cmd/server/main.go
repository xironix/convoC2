package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cxnturi0n/convoC2/pkg/server"
	"github.com/cxnturi0n/convoC2/pkg/server/tui"
)

func init() {
	flag.IntVar(&server.MsgTimeout, "t", 30, "")
	flag.StringVar(&server.BindIp, "b", "0.0.0.0", "")

	flag.IntVar(&server.MsgTimeout, "msgTimeout", 30, "")
	flag.StringVar(&server.BindIp, "bindIp", "0.0.0.0", "")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of convoC2 server:\n")
		fmt.Fprintf(os.Stderr, "  -t, --msgTimeout  How much to wait for command output (default 30 s)\n")
		fmt.Fprintf(os.Stderr, "  -b, --bindIp      Bind IP address (default 0.0.0.0)\n")
	}

	flag.Parse()

	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
