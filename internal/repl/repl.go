package repl

import (
	"fmt"
	"os"

	"github.com/artilugio0/replit"
	tea "github.com/charmbracelet/bubbletea"
)

func initialModel(dbFile string) *replit.REPL {
	suggestions := []string{
		"q.",
		"q.timestamp.gt('1m')",
		"q.id.eq('",
		"q.method.eq('POST')",
		"q.method.eq('GET')",
		"q.method.eq('",
		"q.path.contains('",
		"q.header('",
		"q.header('host').eq('",
		"q.header('content-type').contains('",
		"q.header('content-type').eq('",
		"q.body.contains('",
		"q.raw.contains('",
		"q.resp_header('",
		"q.resp_header('content-type').contains('",
		"q.resp_header('content-type').eq('",
		"q.resp_body.contains('",
		"q.resp_raw.contains('",
	}

	ev := newLuaEvaluator(dbFile)

	repl := replit.NewREPL(ev, replit.WithPromptInitialSuggestions(suggestions))
	ev.repl = repl

	return repl
}

func Run(dbFile string) {
	p := tea.NewProgram(initialModel(dbFile), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
