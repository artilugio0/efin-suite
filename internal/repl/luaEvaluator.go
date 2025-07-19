package repl

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/artilugio0/efin-suite/internal/ql"
	"github.com/artilugio0/efin-testifier/pkg/liblua"
	"github.com/artilugio0/replit"
	lua "github.com/yuin/gopher-lua"
)

//go:embed lua/queryDSL.lua
var queryDSLSource string

type luaEvaluator struct {
	l      *lua.LState
	repl   *replit.REPL
	dbFile string
}

func newLuaEvaluator(dbFile string) *luaEvaluator {
	L := lua.NewState()
	L.OpenLibs()
	liblua.RegisterCommonRuntimeFunctions(L, 20)

	if err := L.DoString(queryDSLSource); err != nil {
		panic(err)
	}

	return &luaEvaluator{
		l:      L,
		dbFile: dbFile,
	}
}

func (le *luaEvaluator) Eval(input string) (*replit.Result, error) {
	ctx := context.TODO()

	value, out, err := execLua(le.l, input)
	if err != nil {
		return nil, err
	}

	if value != nil {
		if mt, ok := le.l.GetMetatable(value).(*lua.LTable); ok {
			if mtID, ok := le.l.GetField(mt, "__mt_id").(lua.LString); ok {
				if mtID == "expr_mt" || mtID == "field_mt" {
					query, err := le.toQuery(value.(*lua.LTable))
					if err != nil {
						return nil, err
					}
					rows, err := doRequestQuery(ctx, le.dbFile, query)
					if err != nil {
						return nil, err
					}

					if len(rows) == 0 {
						return &replit.Result{
							Output: "0 requests found",
						}, nil
					}

					width, height := le.repl.GetWidth(), le.repl.GetHeight()

					return &replit.Result{
						View: NewQueryResultsView(le.dbFile, le.l, rows, width, height),
					}, nil
				}
			}
		}

		out += le.l.ToStringMeta(value).String()
	}

	return &replit.Result{
		Output: out,
	}, nil
}

func execLua(L *lua.LState, code string) (lua.LValue, string, error) {
	oldTop := L.GetTop()
	defer L.SetTop(oldTop)

	prints := []string{}
	oldPrint := L.GetGlobal("print")
	printFunc := func(ls *lua.LState) int {
		n := ls.GetTop()
		args := []string{}
		for i := 1; i <= n; i++ {
			args = append(args, ls.ToStringMeta(ls.Get(i)).String())
		}
		prints = append(prints, strings.Join(args, "\t"))
		return 0
	}
	L.SetGlobal("print", L.NewFunction(printFunc))
	defer L.SetGlobal("print", oldPrint)

	var f *lua.LFunction
	f, err := L.LoadString("return " + code)
	if err != nil {
		f, err = L.LoadString(code)
		if err != nil {
			return nil, "", err
		}
	}
	L.Push(f)

	// Function is now on the stack
	if err := L.PCall(0, lua.MultRet, nil); err != nil {
		return nil, "", err
	}

	// Results are now on the stack starting after oldTop
	newTop := L.GetTop()
	nres := newTop - oldTop
	if nres == 0 {
		return nil, strings.Join(prints, "\n"), nil
	}
	resultValue := L.Get(-1)

	// Clean up results from stack (defer will handle, but explicit)
	L.SetTop(oldTop)

	return resultValue, strings.Join(prints, "\n"), nil
}

func (le *luaEvaluator) toQuery(t *lua.LTable) (*ql.Query, error) {
	query := &ql.Query{
		Operation: ql.QueryOperationGet,
	}
	_ = query

	condition, err := toRequestCondition(t)
	if err != nil {
		return nil, err
	}

	query.RequestCondition = condition
	return query, nil
}

func toRequestCondition(t *lua.LTable) (ql.RequestCondition, error) {
	op := t.RawGet(lua.LString("op"))
	if op == lua.LNil {
		return nil, fmt.Errorf("the query is missing an operation")
	}

	if op.String() == "and" {
		left, ok := t.RawGet(lua.LString("left")).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("invalid left operand for and")
		}

		right, ok := t.RawGet(lua.LString("right")).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("invalid right operand for and")
		}

		leftCond, err := toRequestCondition(left)
		if err != nil {
			return nil, err
		}

		rightCond, err := toRequestCondition(right)
		if err != nil {
			return nil, err
		}

		return &ql.AndCondition{Condition1: leftCond, Condition2: rightCond}, nil
	}

	if op.String() == "or" {
		left, ok := t.RawGet(lua.LString("left")).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("invalid left operand for or")
		}

		right, ok := t.RawGet(lua.LString("right")).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("invalid right operand for or")
		}

		leftCond, err := toRequestCondition(left)
		if err != nil {
			return nil, err
		}

		rightCond, err := toRequestCondition(right)
		if err != nil {
			return nil, err
		}

		return &ql.OrCondition{Condition1: leftCond, Condition2: rightCond}, nil
	}

	if op.String() == "not" {
		expr, ok := t.RawGet(lua.LString("expr")).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("invalid expression for not")
		}

		exprCond, err := toRequestCondition(expr)
		if err != nil {
			return nil, err
		}

		return &ql.NotCondition{Condition: exprCond}, nil
	}

	field := t.RawGet(lua.LString("field"))
	if field == lua.LNil {
		return nil, fmt.Errorf("the query is missing a field")
	}

	value := t.RawGet(lua.LString("value"))
	if value == lua.LNil {
		return nil, fmt.Errorf("the query is missing a value")
	}

	switch field.String() {
	case "id":
		return &ql.RequestIdCondition{Id: value.String(), Operator: op.String()}, nil

	case "timestamp":
		end := len(value.String()) - 1
		vstr := value.String()
		d, err := strconv.Atoi(vstr[:end])
		if err != nil {
			return nil, fmt.Errorf("invalid duration value '%s'", value.String())
		}

		tk := time.Duration(d)
		switch vstr[end] {
		case 's':
			tk *= time.Second
		case 'm':
			tk *= time.Minute
		case 'h':
			tk *= time.Hour
		case 'd':
			tk *= time.Hour * 24
		default:
			return nil, fmt.Errorf("invalid duration suffix: '%c'", vstr[end])
		}

		return &ql.RequestTimestampCondition{TimeAgo: tk, Operator: op.String()}, nil

	case "path":
		return &ql.RequestPathCondition{Value: value.String(), Operator: op.String()}, nil

	case "method":
		return &ql.RequestMethodCondition{Value: strings.ToUpper(value.String())}, nil

	case "body":
		return &ql.RequestBodyCondition{Value: value.String(), Operator: op.String()}, nil

	case "raw":
		return &ql.RequestRawCondition{
			UniqueID: strconv.Itoa(rand.Int()),
			Value:    value.String(),
		}, nil

	case "resp_status":
		status, err := strconv.Atoi(value.String())
		if err != nil {
			return nil, fmt.Errorf("invalid status value '%s'", value.String())
		}
		return &ql.RequestResponseStatusCondition{Value: status, Operator: op.String()}, nil
	case "resp_body":
		return &ql.RequestResponseBodyCondition{Value: value.String(), Operator: op.String()}, nil

	case "resp_raw":
		return &ql.RequestResponseRawCondition{
			UniqueID: strconv.Itoa(rand.Int()),
			Value:    value.String(),
		}, nil

	default:
		if strings.HasPrefix(field.String(), "header:") {
			name := strings.TrimPrefix(field.String(), "header:")
			return &ql.RequestHeaderCondition{
				UniqueID: strconv.Itoa(rand.Int()),
				Name:     name,
				Operator: op.String(),
				Value:    value.String(),
			}, nil
		}

		if strings.HasPrefix(field.String(), "resp_header:") {
			name := strings.TrimPrefix(field.String(), "resp_header:")
			return &ql.RequestResponseHeaderCondition{
				UniqueID: strconv.Itoa(rand.Int()),
				Name:     name,
				Operator: op.String(),
				Value:    value.String(),
			}, nil
		}

		return nil, fmt.Errorf("invalid field '%s'", op.String())
	}
}
