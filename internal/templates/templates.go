package templates

import (
	_ "embed"
)

//go:embed files/request.tpl.py
var pythonScript string

func GetRequestPythonScript() string {
	return pythonScript
}
