package repl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/artilugio0/replit"
	tea "github.com/charmbracelet/bubbletea"
)

func initialModel(dbFile string) *replit.REPL {
	suggestions := []string{
		"proxy connect",
		"proxy config",
		"proxy config set scope ^",
		"proxy config set dbfile ",
		"proxy config set savedir ",
		"query requests where ",
		"query requests where timestamp > 1m ago",
		"query requests where id = ",
		"query requests where method = GET",
		"query requests where method = POST",
		"query requests where method = ",
		"query requests where path contains '",
		"query requests where path = '",
		"query requests where header.",
		"query requests where header.host = '",
		"query requests where header.content-type contains '",
		"query requests where header.content-type = '",
		"query requests where body contains '",
		"query requests where raw contains '",
		"query requests where resp.status = ",
		"query requests where resp.header.",
		"query requests where resp.header.content-type contains '",
		"query requests where resp.header.content-type = '",
		"query requests where resp.raw contains '",
	}

	ev := &evaluator{
		appState: &appState{},
		dbFile:   dbFile,
	}

	repl := replit.NewREPL(ev, replit.WithPromptInitialSuggestions(suggestions))
	ev.repl = repl

	return repl
}

type evaluator struct {
	appState *appState

	repl *replit.REPL

	dbFile string
}

func (e *evaluator) Eval(input string) (*replit.Result, error) {
	ctx := context.TODO()
	fields := strings.Fields(input)

	if len(fields) == 0 {
		return nil, fmt.Errorf("invalid command")
	}

	switch fields[0] {
	case "proxy":
		switch {
		case len(fields) == 2 && fields[1] == "connect":
			return proxyConnectCmd(ctx, e.appState)
		case len(fields) == 2 && fields[1] == "disconnect":
			return proxyDisconnectCmd(e.appState)
		case len(fields) >= 2 && fields[1] == "config":
			if len(fields) == 2 || len(fields) == 3 && fields[2] == "get" {
				return proxyGetConfigCmd(ctx, e.appState)
			}

			if len(fields) >= 5 && fields[2] == "set" {
				return proxySetConfigCmd(ctx, e.appState, fields[3:])
			}

			if len(fields) == 4 && fields[2] == "unset" {
				return proxyUnsetConfigCmd(ctx, e.appState, fields[3])
			}
		}
	case "query":
		return queryCmd(ctx, e.dbFile, input, e.repl.GetWidth(), e.repl.GetHeight())
	}

	return nil, fmt.Errorf("invalid command")
}

func Run(dbFile string) {
	p := tea.NewProgram(initialModel(dbFile), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
