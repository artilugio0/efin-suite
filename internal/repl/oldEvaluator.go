package repl

import (
	"context"
	"fmt"
	"strings"

	"github.com/artilugio0/replit"
)

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
