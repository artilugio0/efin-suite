package templates

import (
	_ "embed"
)

//go:embed files/request.tpl.py
var pythonScript string

//go:embed files/request.tpl.lua
var testifierScript string

func GetRequestPythonScript() string {
	return pythonScript
}

func GetRequestTestifierScript() string {
	return testifierScript
}
