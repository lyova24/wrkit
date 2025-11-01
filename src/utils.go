package src

import (
	"bytes"
	"fmt"
	"text/template"
)

func renderTemplate(tmpl string, vars map[string]string) (string, error) {
	// use text/template, with missingkey=zero (so missing keys produce empty string)
	tpl, err := template.New("cmd").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("bad template %q: %w", tmpl, err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("render template %q: %w", tmpl, err)
	}
	return buf.String(), nil
}

func sampleYAML() string {
	return `# wrkit.yaml — example for wrkit
vars:
  BIN: "./bin/app"

tasks:
  build:
    desc: "Собрать бинарник"
    cmds:
      - "go build -o {{.BIN}} ./..."
    deps: []
    env:
      CGO_ENABLED: "0"
    parallel: false

  test:
    desc: "Запустить тесты"
    cmds:
      - "go test ./..."
    deps: []
    parallel: true

  ci:
    desc: "CI pipeline: тесты -> сборка"
    cmds:
      - "echo CI finished"
    deps:
      - test
      - build
    parallel: false
`
}
