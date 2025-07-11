package repl

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/artilugio0/efin-suite/internal/ql"
	"github.com/artilugio0/efin-suite/internal/templates"
	"github.com/artilugio0/replit"
	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"
)

func queryCmd(ctx context.Context, dbFile string, queryStr string, width, height int) (*replit.Result, error) {
	query, err := ql.ParseQuery(queryStr)
	if err != nil {
		return nil, err
	}

	rows, err := doRequestQuery(ctx, dbFile, query)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &replit.Result{
			Output: "0 requests found",
		}, nil
	}

	return &replit.Result{
		View: NewQueryResultsView(dbFile, rows, width, height),
	}, nil
}

func doRequestQuery(ctx context.Context, dbFile string, query *ql.Query) ([]RequestsTableRow, error) {
	compiled, values, err := query.Compile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dbFile); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// TODO: verify why ctx is being ignored
	rows, err := db.QueryContext(ctx, compiled, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []RequestsTableRow{}

	for rows.Next() {
		var timestamp, requestId, method, url string
		var status int
		if err := rows.Scan(&timestamp, &requestId, &method, &status, &url); err != nil {
			return nil, err
		}
		result = append(result, RequestsTableRow{timestamp, requestId, method, strconv.Itoa(status), url})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type QueryResultsView struct {
	requestsTableView *RequestsTableView
	dbFile            string
	queryRunning      bool
}

func NewQueryResultsView(dbFile string, rows []RequestsTableRow, width, height int) *QueryResultsView {
	requestsTable := NewRequestsTableView(width, height)
	requestsTable.SetRows([]RequestsTableRow(rows))
	requestsTable.SetUpdateFns(func(r RequestsTableRow) string {
		req, err := getRequest(dbFile, r[1])
		if err != nil {
			return fmt.Sprintf("Error getting request: %v", err)
		}

		return rawRequestString(req)
	}, func(r RequestsTableRow) string {
		resp, err := getResponse(dbFile, r[1])
		if err != nil {
			return fmt.Sprintf("Error getting response: %v", err)
		}

		return rawResponseString(resp)
	})

	requestsTable.SetRowKeyBinding("enter", func(row RequestsTableRow) tea.Cmd {
		output := requestsTable.TableRawView()

		return func() tea.Msg {
			req, resp, err := getRequestResponse(dbFile, row[1])
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			if req != nil {
				output += "\n\nRequest ID: " + req.ID
				output += "\n\n" + rawRequestString(req)
			} else {
				output += "\n\n\nRequest not found"
			}

			if resp != nil {
				output += "\n\n\nResponse:"
				output += "\n\n" + rawResponseString(resp)
			} else {
				output += "\n\n\nResponse not found"
			}

			return replit.ExitView{
				Output: output,
			}
		}
	})

	requestsTable.SetRowKeyBinding("s", func(row RequestsTableRow) tea.Cmd {
		reqId := row[1]

		return func() tea.Msg {
			req, err := getRequest(dbFile, reqId)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			scriptTpl := templates.GetRequestPythonScript()
			t, err := template.New("make_request").Parse(scriptTpl)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			f, err := os.OpenFile(req.ID+".py", os.O_RDWR|os.O_CREATE, 0700)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			err = t.Execute(f, req)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}
			if err := f.Close(); err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			return requestTableViewMessage{
				message: "script file saved to " + req.ID + ".py",
			}
		}
	})

	requestsTable.SetRowKeyBinding("t", func(row RequestsTableRow) tea.Cmd {
		reqId := row[1]

		return func() tea.Msg {
			req, err := getRequest(dbFile, reqId)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			funcs := map[string]any{
				"contains": strings.Contains,
				"contains_bytes": func(s []byte, c string) bool {
					return bytes.Contains(s, []byte(c))
				},
			}

			scriptTpl := templates.GetRequestTestifierScript()
			t, err := template.New("make_request").Funcs(funcs).Parse(scriptTpl)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			f, err := os.OpenFile(req.ID+".lua", os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			err = t.Execute(f, req)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}
			if err := f.Close(); err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			return requestTableViewMessage{
				message: "script file saved to " + req.ID + ".lua",
			}
		}
	})

	requestsTable.SetRowKeyBinding("c", func(row RequestsTableRow) tea.Cmd {
		reqId := row[1]

		return func() tea.Msg {
			req, err := getRequest(dbFile, reqId)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			funcs := map[string]any{
				"contains": strings.Contains,
				"contains_bytes": func(s []byte, c string) bool {
					return bytes.Contains(s, []byte(c))
				},
			}

			scriptTpl := templates.GetRequestTestifierScript()
			t, err := template.New("make_request").Funcs(funcs).Parse(scriptTpl)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			f := &strings.Builder{}

			if err := t.Execute(f, req); err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			if err := copyToClipboard(f.String()); err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			return requestTableViewMessage{
				message: "script copied to clipboard",
			}
		}
	})

	requestsTable.SetRowKeyBinding("r", func(row RequestsTableRow) tea.Cmd {
		reqId := row[1]
		output := requestsTable.TableRawView()

		return func() tea.Msg {
			req, err := getRequest(dbFile, reqId)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			if req != nil {
				output += "\n\nRequest ID: " + req.ID
				output += "\n\n" + rawRequestString(req)
			} else {
				output += "\n\n\nRequest not found"
			}

			return replit.ExitView{
				Output: output,
			}
		}
	})

	requestsTable.SetRowKeyBinding("R", func(row RequestsTableRow) tea.Cmd {
		reqId := row[1]
		output := requestsTable.TableRawView()

		return func() tea.Msg {
			_, resp, err := getRequestResponse(dbFile, reqId)
			if err != nil {
				return replit.ExitView{
					Error: err,
				}
			}

			if resp != nil {
				output += "\n\nResponse ID: " + resp.ID
				output += "\n\n" + rawResponseString(resp)
			} else {
				output += "\n\n\nResponse not found"
			}

			return replit.ExitView{
				Output: output,
			}
		}
	})

	return &QueryResultsView{
		requestsTableView: requestsTable,
		dbFile:            dbFile,
		queryRunning:      false,
	}
}

func (v *QueryResultsView) View() string {
	output := v.requestsTableView.View()
	if v.queryRunning {
		output += "\ngetting request data..."
	}
	return output
}

func (v *QueryResultsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := v.requestsTableView.Update(msg)
	v.requestsTableView = m.(*RequestsTableView)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			output := v.requestsTableView.TableRawView()
			return v, func() tea.Msg {
				return replit.ExitView{
					Output: output,
				}
			}
		}
	}

	return v, cmd
}

func (v *QueryResultsView) Init() tea.Cmd {
	return nil
}

func getRequest(dbFile, id string) (*requestEntry, error) {
	if _, err := os.Stat(dbFile); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	reqQuery := `
		SELECT req.timestamp, req.request_id, req.method, req.url, req.body
		FROM requests req
		WHERE req.request_id = ?
	`
	reqRow := db.QueryRow(reqQuery, id)
	if err := reqRow.Err(); err != nil {
		return nil, err
	}
	req := requestEntry{}
	if err := reqRow.Scan(&req.Timestamp, &req.ID, &req.Method, &req.URL, &req.Body); err != nil {
		return nil, err
	}

	reqHeadersQuery := `
		SELECT h.name, h.value
		FROM headers h
		WHERE h.request_id = ?
	`
	reqHeadersRow, err := db.Query(reqHeadersQuery, id)
	if err != nil {
		return nil, err
	}
	defer reqHeadersRow.Close()

	// TODO: add concurrency to speed up this function
	reqHeaders := []headerEntry{}
	for reqHeadersRow.Next() {
		h := headerEntry{}
		if err := reqHeadersRow.Scan(&h.Name, &h.Value); err != nil {
			return nil, err
		}
		reqHeaders = append(reqHeaders, h)
		if strings.ToLower(h.Name) == "host" {
			req.Host = h.Value
		}
	}
	if err := reqHeadersRow.Err(); err != nil {
		return nil, err
	}
	req.Headers = reqHeaders

	return &req, nil
}

func getResponse(dbFile, id string) (*responseEntry, error) {
	if _, err := os.Stat(dbFile); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	respQuery := `
		SELECT resp.response_id, resp.status_code, resp.body
		FROM responses resp
		WHERE resp.response_id = ?
	`
	respRow := db.QueryRow(respQuery, id)
	if err := respRow.Err(); err != nil {
		return nil, err
	}
	resp := responseEntry{}
	if err := respRow.Scan(&resp.ID, &resp.Status, &resp.Body); err != nil {
		return nil, err
	}

	respHeadersQuery := `
		SELECT h.name, h.value
		FROM headers h
		WHERE h.response_id = ?
	`
	respHeadersRows, err := db.Query(respHeadersQuery, id)
	if err != nil {
		return nil, err
	}
	defer respHeadersRows.Close()

	respHeaders := []headerEntry{}
	for respHeadersRows.Next() {
		h := headerEntry{}
		if err := respHeadersRows.Scan(&h.Name, &h.Value); err != nil {
			return nil, err
		}
		respHeaders = append(respHeaders, h)
	}
	if err := respHeadersRows.Err(); err != nil {
		return nil, err
	}
	resp.Headers = respHeaders

	return &resp, nil
}

func getRequestResponse(dbFile, id string) (*requestEntry, *responseEntry, error) {
	req, err := getRequest(dbFile, id)
	if err != nil {
		return nil, nil, err
	}
	resp, err := getResponse(dbFile, id)
	if err != nil {
		return nil, nil, err
	}

	return req, resp, nil
}

type requestEntry struct {
	ID        string
	Timestamp string
	Host      string
	Method    string
	URL       string
	Body      []byte
	Headers   []headerEntry
}

type responseEntry struct {
	ID      string
	Status  int
	Body    []byte
	Headers []headerEntry
}

type headerEntry struct {
	Name  string
	Value string
}

func rawRequestString(req *requestEntry) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %s HTTP/1.1\n", req.Method, req.URL))

	for _, h := range req.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", h.Name, h.Value))
	}

	buf.WriteString("\n")
	buf.Write(req.Body)

	return string(buf.Bytes())
}

func rawResponseString(resp *responseEntry) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\n", resp.Status, http.StatusText(resp.Status)))

	for _, h := range resp.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", h.Name, h.Value))
	}

	buf.WriteString("\n")
	buf.Write(resp.Body)

	return string(buf.Bytes())
}
