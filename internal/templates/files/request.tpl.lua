request = {
  url = 'https://{{ .Host }}{{ .URL }}',
  method = '{{ .Method }}',
  headers = {
  {{- range .Headers }}
    {{- if not (contains .Value "'") }}
    ['{{ .Name }}'] = '{{ .Value }}',
    {{- else if not (contains .Value "\"") }}
    ['{{ .Name }}'] = "{{ .Value }}",
    {{- else }}
    ['{{ .Name }}'] = [==[{{ .Value }}]==],
    {{- end }}
  {{- end }}
  },
  {{- if .Body }}
    {{- if contains_bytes .Body "\n" }}
  body = [==[{{ printf "%s" .Body }}]==],
    {{- else if not (contains_bytes .Body "'") }}
  body = '{{ printf "%s" .Body }}',
    {{- else if not (contains_bytes .Body "\"") }}
  body = "{{ printf "%s" .Body }}",
    {{- else }}
  body = [==[{{ printf "%s" .Body }}]==],
    {{- end }}
  {{- end }}
}
