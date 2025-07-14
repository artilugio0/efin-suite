package ql

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var identifierRE *regexp.Regexp = regexp.MustCompile("^[a-zA-Z_]+$")
var durationRE *regexp.Regexp = regexp.MustCompile("^[1-9][0-9]*[mshd]$")
var uuidRE *regexp.Regexp = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
var headerNameRE *regexp.Regexp = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-]*$")
var numberRE *regexp.Regexp = regexp.MustCompile("^(0|[1-9][0-9]+)$")

type Token interface{}

type Tokenizer struct {
	input string
	pos   int
}

type TokenKeyword string
type TokenOrderOp string
type TokenIdentifier string
type TokenUUID string
type TokenDuration time.Duration
type TokenEOF string
type TokenDotOp string
type TokenExistsOp string
type TokenContainsOp string
type TokenIContainsOp string
type TokenHeaderName string
type TokenString string
type TokenNumber int
type TokenParen string
type TokenLogicalOp string

const (
	TokenKeywordQuery     TokenKeyword     = "query"
	TokenKeywordWhere     TokenKeyword     = "where"
	TokenKeywordAgo       TokenKeyword     = "ago"
	TokenKeywordRequests  TokenKeyword     = "requests"
	TokenKeywordResponses TokenKeyword     = "responses"
	TokenKeywordCount     TokenKeyword     = "count"
	TokenOrderOpEq        TokenOrderOp     = "eq"
	TokenOrderOpGt        TokenOrderOp     = "gt"
	TokenOrderOpGte       TokenOrderOp     = "ge"
	TokenOrderOpLt        TokenOrderOp     = "lt"
	TokenOrderOpLte       TokenOrderOp     = "le"
	TokenDot              TokenDotOp       = "dot"
	TokenExists           TokenExistsOp    = "exists"
	TokenContains         TokenContainsOp  = "contains"
	TokenIContains        TokenIContainsOp = "icontains"
	TokenParenOpen        TokenParen       = "("
	TokenParenClose       TokenParen       = ")"
	TokenLogicalOpAnd     TokenLogicalOp   = "and"
	TokenLogicalOpOr      TokenLogicalOp   = "or"
	TokenLogicalOpNot     TokenLogicalOp   = "not"
	TEOF                  TokenEOF         = ""
)

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: input,
		pos:   0,
	}
}

func (t *Tokenizer) NextToken() (Token, error) {
	token, newPos, err := t.peekToken()
	if err != nil {
		return nil, err
	}

	t.pos = newPos
	return token, nil
}

func (t *Tokenizer) GetPosition() int {
	return t.pos
}

func (t *Tokenizer) SetPosition(pos int) {
	t.pos = pos
}

func (t *Tokenizer) peekToken() (Token, int, error) {
	pos := t.pos
	for pos < len(t.input) && unicode.IsSpace(rune(t.input[pos])) {
		pos++
	}
	start := pos

	isSpecialChar := map[rune]bool{
		'.':  true,
		'"':  true,
		'\'': true,
		'(':  true,
		')':  true,
	}
F:
	for pos < len(t.input) {
		c := rune(t.input[pos])
		if unicode.IsSpace(c) || isSpecialChar[c] {
			break
		}

		switch c {
		case '=':
			pos++
			break F
		case '<', '>':
			pos++
			if pos < len(t.input) && t.input[pos] == '=' {
				pos++
			}
			break F
		}

		pos++
	}

	s := strings.ToLower(t.input[start:pos])
	if s == "" {
		if pos >= len(t.input) {
			return TEOF, pos, nil
		}

		switch t.input[pos] {
		case '.':
			return TokenDot, pos + 1, nil
		case '"', '\'':
			return t.peekStringToken(pos)
		case '(', ')':
			return TokenParen(t.input[pos]), pos + 1, nil
		}
	}

	switch s {
	case "query":
		return TokenKeywordQuery, pos, nil
	case "requests":
		return TokenKeywordRequests, pos, nil
	case "where":
		return TokenKeywordWhere, pos, nil
	case "=":
		return TokenOrderOpEq, pos, nil
	case ">":
		return TokenOrderOpGt, pos, nil
	case ">=":
		return TokenOrderOpGte, pos, nil
	case "<":
		return TokenOrderOpLt, pos, nil
	case "<=":
		return TokenOrderOpLte, pos, nil
	case "ago":
		return TokenKeywordAgo, pos, nil
	case "exists":
		return TokenExists, pos, nil
	case "contains":
		return TokenContains, pos, nil
	case "icontains":
		return TokenIContains, pos, nil
	case "and":
		return TokenLogicalOpAnd, pos, nil
	case "or":
		return TokenLogicalOpOr, pos, nil
	case "not":
		return TokenLogicalOpNot, pos, nil
	}

	if identifierRE.MatchString(s) {
		return TokenIdentifier(t.input[start:pos]), pos, nil
	}

	if durationRE.MatchString(s) {
		d, err := strconv.Atoi(t.input[start : pos-1])
		if err != nil {
			return nil, 0, err
		}
		tk := time.Duration(d)
		switch t.input[pos-1] {
		case 's':
			tk *= time.Second
		case 'm':
			tk *= time.Minute
		case 'h':
			tk *= time.Hour
		case 'd':
			tk *= time.Hour * 24
		default:
			return nil, 0, fmt.Errorf("invalid duration suffix: '%c'", t.input[pos])
		}

		return TokenDuration(tk), pos, nil
	}

	if uuidRE.MatchString(s) {
		return TokenUUID(t.input[start:pos]), pos, nil
	}

	if headerNameRE.MatchString(s) {
		return TokenHeaderName(t.input[start:pos]), pos, nil
	}

	if numberRE.MatchString(s) {
		n, err := strconv.Atoi(t.input[start:pos])
		if err != nil {
			return nil, 0, fmt.Errorf("invalid number: '%s'", t.input[start:pos])
		}
		return TokenNumber(n), pos, nil
	}

	return nil, 0, fmt.Errorf("invalid token found: '%s'", t.input[start:pos])
}

func (t *Tokenizer) peekStringToken(start int) (Token, int, error) {
	if start >= len(t.input) {
		return nil, 0, fmt.Errorf("unexpected end of input")
	}

	if t.input[start] != '"' && t.input[start] != '\'' {
		return nil, 0, fmt.Errorf(`expected a single or double quote. Got '%c'`, t.input[start])
	}

	quote := t.input[start]

	pos := start + 1
	for pos < len(t.input) && t.input[pos] != quote {
		pos++
	}

	if pos >= len(t.input) {
		return nil, 0, fmt.Errorf("unexpected end of input")
	}

	return TokenString(t.input[start+1 : pos]), pos + 1, nil
}

func (t *Tokenizer) PeekToken() (Token, error) {
	token, _, err := t.peekToken()
	return token, err
}

func (t *Tokenizer) AssertNextToken(expectedToken Token) error {
	start := t.pos
	nt, err := t.NextToken()
	if err != nil {
		return err
	}

	if nt == TEOF && nt != expectedToken {
		t.pos = start
		return fmt.Errorf("expected token '%s'. Found end of input instead", expectedToken)
	}

	if nt != expectedToken {
		t.pos = start
		if nt == TEOF {
			return fmt.Errorf("expected token '%s'. Found end of input instead", expectedToken)
		}
		return fmt.Errorf("unexpected token found: '%s'", nt)
	}

	return nil
}

func (t *Tokenizer) AssertNextTokenOneOf(alts []Token) error {
	var err error
	for _, alt := range alts {
		if err = t.AssertNextToken(alt); err == nil {
			return nil
		}
	}

	return err
}

func NextTokenWithType[A any](tokenizer *Tokenizer) (*A, error) {
	start := tokenizer.GetPosition()
	nt, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	if token, ok := nt.(A); ok {
		return &token, nil
	}

	tokenizer.SetPosition(start)

	var zeroValue A
	typ := reflect.TypeOf(zeroValue)
	typeFields := strings.Split(typ.String(), ".")
	typeStr := typeFields[len(typeFields)-1]
	return nil, fmt.Errorf("expected a %s. Found '%s' instead", typeStr, nt)
}

func ParseQuery(s string) (*Query, error) {
	tk := NewTokenizer(s)
	result := &Query{}

	err := tk.AssertNextToken(TokenKeywordQuery)
	if err != nil {
		return nil, err
	}

	token, err := tk.NextToken()
	if err != nil {
		return nil, err
	}
	switch token {
	case TokenKeywordRequests:
		err := tk.AssertNextToken(TokenKeywordWhere)
		if err != nil {
			return nil, err
		}

		result.Operation = QueryOperationGet
		cond, err := parseRequestCondition(tk, true)
		if err != nil {
			return nil, err
		}
		result.RequestCondition = cond

		if err := tk.AssertNextToken(TEOF); err != nil {
			return nil, err
		}

	default:
		if token == TEOF {
			return nil, fmt.Errorf("expected 'requests', 'responses' or 'count'. Found input end instead")
		}

		return nil, fmt.Errorf("Expected 'requests', 'responses' or 'count'. Found '%s' instead", token)
	}

	return result, nil
}

func parseRequestCondition(tokenizer *Tokenizer, andOr bool) (RequestCondition, error) {
	nt, err := tokenizer.PeekToken()
	if err != nil {
		return nil, err
	}

	var cond RequestCondition
	switch nt {
	case TokenParenOpen:
		if err := tokenizer.AssertNextToken(TokenParenOpen); err != nil {
			return nil, err
		}

		cond, err := parseRequestCondition(tokenizer, true)
		if err != nil {
			return nil, err
		}

		if err := tokenizer.AssertNextToken(TokenParenClose); err != nil {
			return nil, err
		}
		return cond, nil

	case TokenIdentifier("timestamp"):
		cond, err = parseRequestTimestampCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("id"):
		cond, err = parseRequestIdCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("method"):
		cond, err = parseRequestMethodCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("header"):
		cond, err = parseRequestHeaderCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("body"):
		cond, err = parseRequestBodyCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("path"):
		cond, err = parseRequestPathCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("raw"):
		cond, err = parseRequestRawCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenIdentifier("resp"), TokenIdentifier("response"):
		cond, err = parseRequestResponseCondition(tokenizer)
		if err != nil {
			return nil, err
		}
	case TokenLogicalOpNot:
		if err := tokenizer.AssertNextToken(TokenLogicalOpNot); err != nil {
			return nil, err
		}

		cond, err = parseRequestCondition(tokenizer, false)
		if err != nil {
			return nil, err
		}

		cond = &NotCondition{
			Condition: cond,
		}
	default:
		if nt == TEOF {
			return nil, fmt.Errorf("expected a request condition. Found input end instead")
		}

		return nil, fmt.Errorf("expected a request condition. Found '%s' instead", nt)
	}

	if !andOr {
		return cond, nil
	}

	nt, err = tokenizer.PeekToken()
	if err != nil {
		return nil, err
	}

	switch nt {
	case TokenLogicalOpAnd:
		if err := tokenizer.AssertNextToken(TokenLogicalOpAnd); err != nil {
			return nil, err
		}

		cond2, err := parseRequestCondition(tokenizer, true)
		if err != nil {
			return nil, err
		}

		cond = &AndCondition{
			Condition1: cond,
			Condition2: cond2,
		}

	case TokenLogicalOpOr:
		if err := tokenizer.AssertNextToken(TokenLogicalOpOr); err != nil {
			return nil, err
		}

		cond2, err := parseRequestCondition(tokenizer, true)
		if err != nil {
			return nil, err
		}

		cond = &OrCondition{
			Condition1: cond,
			Condition2: cond2,
		}
	}

	return cond, nil
}

func parseRequestTimestampCondition(tokenizer *Tokenizer) (*RequestTimestampCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("timestamp")); err != nil {
		return nil, err
	}

	orderOp, err := NextTokenWithType[TokenOrderOp](tokenizer)
	if err != nil {
		return nil, err
	}
	var operator string
	switch *orderOp {
	case TokenOrderOpEq:
		operator = "eq"
	case TokenOrderOpLt:
		operator = "lt"
	case TokenOrderOpLte:
		operator = "le"
	case TokenOrderOpGt:
		operator = "gt"
	case TokenOrderOpGte:
		operator = "ge"
	default:
		return nil, fmt.Errorf("unknown timestamp operator: '%s'", *orderOp)
	}

	duration, err := NextTokenWithType[TokenDuration](tokenizer)
	if err != nil {
		return nil, err
	}

	if err := tokenizer.AssertNextToken(TokenKeywordAgo); err != nil {
		return nil, err
	}

	return &RequestTimestampCondition{
		TimeAgo:  time.Duration(*duration),
		Operator: operator,
	}, nil
}

func parseRequestIdCondition(tokenizer *Tokenizer) (*RequestIdCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("id")); err != nil {
		return nil, err
	}

	orderOp, err := NextTokenWithType[TokenOrderOp](tokenizer)
	if err != nil {
		return nil, err
	}
	var operator string
	switch *orderOp {
	case TokenOrderOpEq:
		operator = "eq"
	default:
		return nil, fmt.Errorf("invalid order operator: '%s'", *orderOp)
	}

	var id string
	idToken, err := NextTokenWithType[TokenUUID](tokenizer)
	if err == nil {
		id = string(*idToken)
	} else {
		// TODO: fix ID condition
		numberToken, err2 := NextTokenWithType[TokenNumber](tokenizer)
		if err2 == nil {
			id = string(*numberToken)
		} else {
			stringToken, err3 := NextTokenWithType[TokenString](tokenizer)
			if err3 != nil {
				return nil, err
			}
			id = string(*stringToken)
		}
	}

	return &RequestIdCondition{
		Id:       id,
		Operator: operator,
	}, nil
}

func parseRequestMethodCondition(tokenizer *Tokenizer) (*RequestMethodCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("method")); err != nil {
		return nil, err
	}

	err := tokenizer.AssertNextToken(TokenOrderOpEq)
	if err != nil {
		return nil, err
	}

	var method string
	nt, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	switch v := nt.(type) {
	case TokenIdentifier:
		method = strings.ToUpper(string(v))
	case TokenString:
		method = strings.ToUpper(string(v))
	default:
		return nil, fmt.Errorf("invalid request method")
	}

	if method != "GET" && method != "POST" && method != "PUT" &&
		method != "PATCH" && method != "DELETE" && method != "CONNECT" &&
		method != "HEAD" && method != "OPTION" {
		return nil, fmt.Errorf("invalid request method")
	}

	return &RequestMethodCondition{
		Value: method,
	}, nil
}

func parseRequestHeaderCondition(tokenizer *Tokenizer) (*RequestHeaderCondition, error) {
	startPosition := tokenizer.GetPosition()

	if err := tokenizer.AssertNextToken(TokenIdentifier("header")); err != nil {
		return nil, err
	}

	if err := tokenizer.AssertNextToken(TokenDot); err != nil {
		return nil, err
	}

	headerNameToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	headerName := ""
	switch v := headerNameToken.(type) {
	case TokenHeaderName:
		headerName = string(v)
	case TokenIdentifier:
		headerName = string(v)
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenExists:
		return &RequestHeaderCondition{
			UniqueID: strconv.Itoa(startPosition),
			Name:     headerName,
			Operator: "exists",
		}, nil

	case TokenOrderOpEq:
		operator = "eq"
	case TokenContains:
		operator = "contains"
	case TokenIContains:
		operator = "icontains"
	default:
		return nil, fmt.Errorf("unknown header operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestHeaderCondition{
		UniqueID: strconv.Itoa(startPosition),
		Name:     headerName,
		Operator: operator,
		Value:    string(*value),
	}, nil
}

func parseRequestBodyCondition(tokenizer *Tokenizer) (*RequestBodyCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("body")); err != nil {
		return nil, err
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenContains:
		operator = "contains"
	case TokenIContains:
		operator = "icontains"
	default:
		return nil, fmt.Errorf("unknown body operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestBodyCondition{
		Operator: operator,
		Value:    string(*value),
	}, nil
}

func parseRequestPathCondition(tokenizer *Tokenizer) (*RequestPathCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("path")); err != nil {
		return nil, err
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenOrderOpEq:
		operator = "eq"
	case TokenContains:
		operator = "contains"
	case TokenIContains:
		operator = "icontains"
	default:
		return nil, fmt.Errorf("unknown path operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestPathCondition{
		Operator: operator,
		Value:    string(*value),
	}, nil
}

func parseRequestRawCondition(tokenizer *Tokenizer) (*RequestRawCondition, error) {
	startPosition := tokenizer.GetPosition()

	if err := tokenizer.AssertNextToken(TokenIdentifier("raw")); err != nil {
		return nil, err
	}

	err := tokenizer.AssertNextToken(TokenContains)
	if err != nil {
		return nil, err
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestRawCondition{
		UniqueID: strconv.Itoa(startPosition),
		Value:    string(*value),
	}, nil
}

func parseRequestResponseCondition(tokenizer *Tokenizer) (RequestCondition, error) {
	if err := tokenizer.AssertNextTokenOneOf([]Token{
		TokenIdentifier("response"),
		TokenIdentifier("resp"),
	}); err != nil {
		return nil, err
	}

	if err := tokenizer.AssertNextToken(TokenDot); err != nil {
		return nil, err
	}

	nt, err := tokenizer.PeekToken()
	if err != nil {
		return nil, err
	}

	responseFieldName, ok := nt.(TokenIdentifier)
	if !ok {
		return nil, fmt.Errorf("invalid response field found")
	}

	switch string(responseFieldName) {
	case "header":
		return parseRequestResponseHeaderCondition(tokenizer)
	case "body":
		return parseRequestResponseBodyCondition(tokenizer)
	case "status":
		return parseRequestResponseStatusCondition(tokenizer)
	case "raw":
		return parseRequestResponseRawCondition(tokenizer)
	}

	return nil, fmt.Errorf("invalid response field '%s'", string(responseFieldName))
}

func parseRequestResponseBodyCondition(tokenizer *Tokenizer) (*RequestResponseBodyCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("body")); err != nil {
		return nil, err
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenContains:
		operator = "contains"
	case TokenIContains:
		operator = "icontains"
	default:
		return nil, fmt.Errorf("unknown body operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestResponseBodyCondition{
		Operator: operator,
		Value:    string(*value),
	}, nil
}

func parseRequestResponseStatusCondition(tokenizer *Tokenizer) (*RequestResponseStatusCondition, error) {
	if err := tokenizer.AssertNextToken(TokenIdentifier("status")); err != nil {
		return nil, err
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenOrderOpEq:
		operator = "eq"
	case TokenOrderOpLt:
		operator = "lt"
	case TokenOrderOpLte:
		operator = "le"
	case TokenOrderOpGt:
		operator = "gt"
	case TokenOrderOpGte:
		operator = "ge"
	default:
		return nil, fmt.Errorf("unknown status operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenNumber](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestResponseStatusCondition{
		Operator: operator,
		Value:    int(*value),
	}, nil
}

func parseRequestResponseHeaderCondition(tokenizer *Tokenizer) (*RequestResponseHeaderCondition, error) {
	startPosition := tokenizer.GetPosition()

	if err := tokenizer.AssertNextToken(TokenIdentifier("header")); err != nil {
		return nil, err
	}

	if err := tokenizer.AssertNextToken(TokenDot); err != nil {
		return nil, err
	}

	headerNameToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	headerName := ""
	switch v := headerNameToken.(type) {
	case TokenHeaderName:
		headerName = string(v)
	case TokenIdentifier:
		headerName = string(v)
	}

	opToken, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	operator := ""
	switch opToken {
	case TokenExists:
		return &RequestResponseHeaderCondition{
			UniqueID: strconv.Itoa(startPosition),
			Name:     headerName,
			Operator: "exists",
		}, nil

	case TokenOrderOpEq:
		operator = "eq"
	case TokenContains:
		operator = "contains"
	case TokenIContains:
		operator = "icontains"
	default:
		return nil, fmt.Errorf("unknown header operator: '%s'", operator)
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestResponseHeaderCondition{
		UniqueID: strconv.Itoa(startPosition),
		Name:     headerName,
		Operator: operator,
		Value:    string(*value),
	}, nil
}

func parseRequestResponseRawCondition(tokenizer *Tokenizer) (*RequestResponseRawCondition, error) {
	startPosition := tokenizer.GetPosition()

	if err := tokenizer.AssertNextToken(TokenIdentifier("raw")); err != nil {
		return nil, err
	}

	err := tokenizer.AssertNextToken(TokenContains)
	if err != nil {
		return nil, err
	}

	value, err := NextTokenWithType[TokenString](tokenizer)
	if err != nil {
		return nil, err
	}

	return &RequestResponseRawCondition{
		UniqueID: strconv.Itoa(startPosition),
		Value:    string(*value),
	}, nil
}
