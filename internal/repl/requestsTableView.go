package repl

import (
	"fmt"

	"github.com/artilugio0/replit"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusTable int = iota
	focusVp1
	focusVp2
)

type RequestsTableRow []string

type RequestsTableView struct {
	width  int
	height int

	rows       []RequestsTableRow
	currentRow int
	table      table.Model

	vp1 *replit.Viewport
	vp2 *replit.Viewport

	updateVp1Fn func(RequestsTableRow) string
	updateVp2Fn func(RequestsTableRow) string

	rowKeyBindings map[string](func(RequestsTableRow) tea.Cmd)

	focus        int
	focusStyle   lipgloss.Style
	unfocusStyle lipgloss.Style
}

func NewRequestsTableView(width, height int) *RequestsTableView {
	vp1 := replit.NewViewport(replit.ShowEmptyLines(true))
	vp1.SetSize(width, height)
	vp2 := replit.NewViewport(replit.ShowEmptyLines(true))
	vp2.SetSize(width, height)

	focusStyle := lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("228")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true)

	unfocusStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true)

	vp1.SetStyle(unfocusStyle)
	vp2.SetStyle(unfocusStyle)

	return &RequestsTableView{
		width:          width,
		height:         height,
		vp1:            vp1,
		vp2:            vp2,
		rowKeyBindings: map[string](func(RequestsTableRow) tea.Cmd){},
		focusStyle:     focusStyle,
		unfocusStyle:   unfocusStyle,
	}
}

func (v *RequestsTableView) SetRows(rows []RequestsTableRow) {
	v.rows = rows
	v.currentRow = 0

	maxColWidths := make([]int, 5)

	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row(r)
		for j, c := range r {
			// max colum size is width / 2
			maxColWidths[j] = max(maxColWidths[j], min(max(10, len(c)), v.width/2))
		}
	}

	maxColWidthsSum := 0
	for _, w := range maxColWidths {
		maxColWidthsSum += w
	}

	columns := []table.Column{
		{Title: "Timestamp", Width: int(float32(maxColWidths[0]) / float32(maxColWidthsSum) * 0.95 * float32(v.width))},
		{Title: "ID", Width: int(float32(maxColWidths[1]) / float32(maxColWidthsSum) * 0.95 * float32(v.width))},
		{Title: "Method", Width: int(float32(maxColWidths[2]) / float32(maxColWidthsSum) * 0.95 * float32(v.width))},
		{Title: "Status", Width: int(float32(maxColWidths[3]) / float32(maxColWidthsSum) * 0.95 * float32(v.width))},
		{Title: "URL", Width: int(float32(maxColWidths[4]) / float32(maxColWidthsSum) * 0.95 * float32(v.width))},
	}

	maxHeight := v.height/2 - 1
	tableHeight := min(maxHeight, len(rows)+1)

	v.vp1.SetSize(v.width/2, v.height-tableHeight-1)
	v.vp2.SetSize(v.width/2+v.width%2, v.height-tableHeight-1)

	v.table = table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
		table.WithWidth(v.width),
	)

	v.updateViewports()
}

func (v *RequestsTableView) updateViewports() {
	if len(v.rows) > 0 {
		if v.updateVp1Fn != nil {
			vp1Content := v.updateVp1Fn(v.rows[v.currentRow])
			v.vp1.Clear()
			v.vp1.AppendBlock(replit.StringBlock{vp1Content})
			v.vp1.GotoTop()
		}
		if v.updateVp2Fn != nil {
			vp2Content := v.updateVp2Fn(v.rows[v.currentRow])
			v.vp2.Clear()
			v.vp2.AppendBlock(replit.StringBlock{vp2Content})
			v.vp2.GotoTop()
		}
	}
}

func (v *RequestsTableView) SetUpdateFns(fn1, fn2 func(RequestsTableRow) string) {
	v.updateVp1Fn = fn1
	v.updateVp2Fn = fn2

	v.updateViewports()
}

func (v *RequestsTableView) SetRowKeyBinding(key string, fn func(RequestsTableRow) tea.Cmd) {
	v.rowKeyBindings[key] = fn
}

func (v *RequestsTableView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "K", "ctrl+k":
			v.focus = focusTable
			v.vp1.SetStyle(v.unfocusStyle)
			v.vp2.SetStyle(v.unfocusStyle)
			return v, nil
		case "J", "ctrl+j":
			if v.focus == focusTable {
				v.focus = focusVp1
				v.vp1.SetStyle(v.focusStyle)
				v.vp2.SetStyle(v.unfocusStyle)
				return v, nil
			}
		case "H", "ctrl+h":
			v.focus = focusVp1
			v.vp1.SetStyle(v.focusStyle)
			v.vp2.SetStyle(v.unfocusStyle)
			return v, nil
		case "L", "ctrl+l":
			v.focus = focusVp2
			v.vp1.SetStyle(v.unfocusStyle)
			v.vp2.SetStyle(v.focusStyle)
			return v, nil
		}
	case tea.WindowSizeMsg:
		v.height = msg.Height
		v.width = msg.Width

		maxHeight := v.height/2 - 1
		tableHeight := min(maxHeight, len(v.rows)+1)
		v.table.SetWidth(v.width)
		v.table.SetHeight(tableHeight)

		v.vp1.SetSize(v.width/2, v.height-tableHeight)
		v.vp2.SetSize(v.width/2+v.width%2, v.height-tableHeight)
	}

	var cmd tea.Cmd
	switch v.focus {
	case focusTable:
		v.table, cmd = v.table.Update(msg)
		selectedRow := v.table.Cursor()
		if selectedRow != v.currentRow {
			v.currentRow = selectedRow
			v.updateViewports()
		}

		if kmsg, ok := msg.(tea.KeyMsg); ok {
			kb, ok := v.rowKeyBindings[kmsg.String()]
			if ok {
				cmd = kb(RequestsTableRow(v.table.SelectedRow()))
			}
		}

	case focusVp1:
		vp, c := v.vp1.Update(msg)
		v.vp1 = vp.(*replit.Viewport)
		cmd = c

	case focusVp2:
		vp, c := v.vp2.Update(msg)
		v.vp2 = vp.(*replit.Viewport)
		cmd = c
	}

	return v, cmd
}

func (v *RequestsTableView) View() string {
	table := v.table.View()
	table += fmt.Sprintf("\n%d requests found", len(v.table.Rows()))

	viewports := lipgloss.JoinHorizontal(lipgloss.Bottom, v.vp1.View(), v.vp2.View())
	output := lipgloss.JoinVertical(lipgloss.Left, table, viewports)

	return output
}

func (v *RequestsTableView) Init() tea.Cmd {
	return nil
}

func (v *RequestsTableView) TableRawView() string {
	output := ""
	rows := v.table.Rows()
	v.table.SetHeight(len(rows) + 1)

	s := table.Styles{
		Header:   lipgloss.NewStyle().Padding(0, 1, 0, 0),
		Cell:     lipgloss.NewStyle().Padding(0, 1, 0, 0),
		Selected: lipgloss.NewStyle(),
	}

	v.table.SetStyles(s)
	output += v.table.View()
	output += fmt.Sprintf("\n%d requests found", len(rows))

	return output
}

// cmdDone is a message indicating that a repl command has finishied
// its execution
type CmdDone struct {
	Output string
	Err    error
}
