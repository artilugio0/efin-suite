package ql

import (
	"fmt"
	"strconv"
	"time"
)

type QueryOperation int

const (
	QueryOperationGet QueryOperation = iota
	QueryOperationCount
)

type Query struct {
	Operation        QueryOperation
	RequestCondition RequestCondition
}

type RequestCondition interface {
	GetRequestConditionString() (string, []any, error)
	GetRequestJoinsString() (string, error)
}

type RequestIdCondition struct {
	Id       string
	Operator string
}

func (c *RequestIdCondition) GetRequestConditionString() (string, []any, error) {
	var op string
	switch c.Operator {
	case "eq":
		op = "="
	default:
		return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
	}

	condition := "req.request_id " + op + " ?"

	return condition, []any{c.Id}, nil
}

func (c *RequestIdCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestMethodCondition struct {
	Value string
}

func (c *RequestMethodCondition) GetRequestConditionString() (string, []any, error) {
	condition := "req.method = ?"
	return condition, []any{c.Value}, nil
}

func (c *RequestMethodCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestPathCondition struct {
	Operator string
	Value    string
}

func (c *RequestPathCondition) GetRequestConditionString() (string, []any, error) {
	switch c.Operator {
	case "eq":
		condition := "req.url = ?"
		return condition, []any{c.Value}, nil

	case "contains":
		condition := "req.url LIKE ?"
		return condition, []any{"%" + c.Value + "%"}, nil

	case "icontains":
		condition := "LOWER(req.url) = LIKE LOWER(?)"
		return condition, []any{"%" + c.Value + "%"}, nil
	}

	return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
}

func (c *RequestPathCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestHeaderCondition struct {
	UniqueID string
	Name     string
	Operator string
	Value    string
}

func (c *RequestHeaderCondition) GetRequestConditionString() (string, []any, error) {
	condition := ""
	tableName := "h" + c.UniqueID

	switch c.Operator {
	case "exists":
		condition = "LOWER(" + tableName + ".name) = LOWER(?)"
		return condition, []any{c.Name}, nil

	case "contains":
		condition = "LOWER(" + tableName + ".name) = LOWER(?) and " + tableName + ".value LIKE ?"
		return condition, []any{c.Name, "%" + c.Value + "%"}, nil

	case "icontains":
		condition = "LOWER(" + tableName + ".name) = LOWER(?) and LOWER(" + tableName + ".value) LIKE LOWER(?)"
		return condition, []any{c.Name, "%" + c.Value + "%"}, nil

	default:
		return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
	}

	return condition, nil, nil
}

func (c *RequestHeaderCondition) GetRequestJoinsString() (string, error) {
	return "INNER JOIN headers h" + c.UniqueID + " ON h" + c.UniqueID + ".request_id = req.request_id", nil
}

type RequestBodyCondition struct {
	Operator string
	Value    string
}

func (c *RequestBodyCondition) GetRequestConditionString() (string, []any, error) {
	switch c.Operator {
	case "contains":
		condition := "req.body LIKE ?"
		return condition, []any{"%" + c.Value + "%"}, nil

	case "icontains":
		condition := "LOWER(req.body) = LIKE LOWER(?)"
		return condition, []any{"%" + c.Value + "%"}, nil
	}

	return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
}

func (c *RequestBodyCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestRawCondition struct {
	UniqueID string
	Value    string
}

func (c *RequestRawCondition) GetRequestConditionString() (string, []any, error) {
	headerTable := "h" + c.UniqueID
	_ = headerTable
	value := "%" + c.Value + "%"

	condition := ("req.url LIKE ?" +
		"or req.method LIKE ? " +
		"or req.body LIKE ? " +
		"or " + headerTable + ".value LIKE ?" +
		"or " + headerTable + ".name LIKE ?")
	return condition, []any{value, value, value, value, value}, nil
}

func (c *RequestRawCondition) GetRequestJoinsString() (string, error) {
	return "INNER JOIN headers h" + c.UniqueID + " ON h" + c.UniqueID + ".request_id = req.request_id", nil
}

type RequestTimestampCondition struct {
	TimeAgo  time.Duration
	Operator string
}

func (c *RequestTimestampCondition) GetRequestConditionString() (string, []any, error) {
	seconds := strconv.Itoa(int(c.TimeAgo.Seconds()))
	var op string
	switch c.Operator {
	case "eq":
		op = "="
	case "lt":
		op = "<"
	case "le":
		op = "<="
	case "gt":
		op = ">"
	case "ge":
		op = ">="
	default:
		return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
	}
	condition := "req.timestamp " + op + " datetime('now', '-" + seconds + " seconds')"

	return condition, nil, nil
}

func (c *RequestTimestampCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestResponseStatusCondition struct {
	Operator string
	Value    int
}

func (c *RequestResponseStatusCondition) GetRequestConditionString() (string, []any, error) {
	op := ""
	switch c.Operator {
	case "eq":
		op = "="
	case "lt":
		op = "<"
	case "le":
		op = "<="
	case "gt":
		op = ">"
	case "ge":
		op = ">="
	default:
		return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
	}
	condition := "resp.status_code " + op + " ?"

	return condition, []any{c.Value}, nil
}

func (c *RequestResponseStatusCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestResponseHeaderCondition struct {
	UniqueID string
	Name     string
	Operator string
	Value    string
}

func (c *RequestResponseHeaderCondition) GetRequestConditionString() (string, []any, error) {
	condition := ""
	tableName := "h" + c.UniqueID

	switch c.Operator {
	case "exists":
		condition = "LOWER(" + tableName + ".name) = LOWER(?)"
		return condition, []any{c.Name}, nil

	case "contains":
		condition = "LOWER(" + tableName + ".name) = LOWER(?) and " + tableName + ".value LIKE ?"
		return condition, []any{c.Name, "%" + c.Value + "%"}, nil

	case "icontains":
		condition = "LOWER(" + tableName + ".name) = LOWER(?) and LOWER(" + tableName + ".value) LIKE LOWER(?)"
		return condition, []any{c.Name, "%" + c.Value + "%"}, nil

	default:
		return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
	}

	return condition, nil, nil
}

func (c *RequestResponseHeaderCondition) GetRequestJoinsString() (string, error) {
	return "INNER JOIN headers h" + c.UniqueID + " ON h" + c.UniqueID + ".response_id = resp.response_id", nil
}

type RequestResponseBodyCondition struct {
	Operator string
	Value    string
}

func (c *RequestResponseBodyCondition) GetRequestConditionString() (string, []any, error) {
	switch c.Operator {
	case "contains":
		condition := "resp.body LIKE ?"
		return condition, []any{"%" + c.Value + "%"}, nil

	case "icontains":
		condition := "LOWER(resp.body) = LIKE LOWER(?)"
		return condition, []any{"%" + c.Value + "%"}, nil
	}

	return "", nil, fmt.Errorf("invalid operator '%s'", c.Operator)
}

func (c *RequestResponseBodyCondition) GetRequestJoinsString() (string, error) {
	return "", nil
}

type RequestResponseRawCondition struct {
	UniqueID string
	Value    string
}

func (c *RequestResponseRawCondition) GetRequestConditionString() (string, []any, error) {
	headerTable := "h" + c.UniqueID
	value := "%" + c.Value + "%"

	condition := ("resp.body LIKE ? " +
		"or " + headerTable + ".value LIKE ?" +
		"or " + headerTable + ".name LIKE ?")
	return condition, []any{value, value, value}, nil
}

func (c *RequestResponseRawCondition) GetRequestJoinsString() (string, error) {
	return "INNER JOIN headers h" + c.UniqueID + " ON h" + c.UniqueID + ".response_id = resp.response_id", nil
}

type NotCondition struct {
	Condition RequestCondition
}

func (c *NotCondition) GetRequestConditionString() (string, []any, error) {
	cond, values, err := c.Condition.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}

	condition := "not (" + cond + ")"
	return condition, values, nil
}

func (c *NotCondition) GetRequestJoinsString() (string, error) {
	joins, err := c.Condition.GetRequestJoinsString()
	if err != nil {
		return "", err
	}

	return joins, nil
}

type AndCondition struct {
	Condition1 RequestCondition
	Condition2 RequestCondition
}

func (c *AndCondition) GetRequestConditionString() (string, []any, error) {
	cond1, values1, err := c.Condition1.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}

	cond2, values2, err := c.Condition2.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}

	condition := "(" + cond1 + ") and (" + cond2 + ")"
	values := append(values1, values2...)

	return condition, values, nil
}

func (c *AndCondition) GetRequestJoinsString() (string, error) {
	joins1, err := c.Condition1.GetRequestJoinsString()
	if err != nil {
		return "", err
	}

	joins2, err := c.Condition2.GetRequestJoinsString()
	if err != nil {
		return "", err
	}

	joins := joins1
	if joins2 != "" {
		joins += " " + joins2
	}

	return joins, nil
}

type OrCondition struct {
	Condition1 RequestCondition
	Condition2 RequestCondition
}

func (c *OrCondition) GetRequestConditionString() (string, []any, error) {
	cond1, values1, err := c.Condition1.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}

	cond2, values2, err := c.Condition2.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}

	condition := "(" + cond1 + ") or (" + cond2 + ")"
	values := append(values1, values2...)

	return condition, values, nil
}

func (c *OrCondition) GetRequestJoinsString() (string, error) {
	joins1, err := c.Condition1.GetRequestJoinsString()
	if err != nil {
		return "", err
	}

	joins2, err := c.Condition2.GetRequestJoinsString()
	if err != nil {
		return "", err
	}

	joins := joins1
	if joins2 != "" {
		joins += " " + joins2
	}

	return joins, nil
}

func (q *Query) Compile() (string, []any, error) {
	conditions, values, err := q.RequestCondition.GetRequestConditionString()
	if err != nil {
		return "", nil, err
	}
	joins, err := q.RequestCondition.GetRequestJoinsString()
	if err != nil {
		return "", nil, err
	}

	query := "SELECT DISTINCT req.timestamp, req.request_id, req.method, resp.status_code, req.url FROM requests req"
	query += " INNER JOIN responses resp on req.request_id = resp.response_id"

	if joins != "" {
		query += " " + joins
	}

	if conditions != "" {
		query += " WHERE " + conditions
	}

	query += " " + "ORDER BY req.timestamp DESC"

	return query, values, nil
}
