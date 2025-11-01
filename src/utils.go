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
  SLEEP_ALL_SUCCESS_MSG: "all sleep tasks executed successfully!"

tasks:
  sleep-for-2:
    desc: "sleep for 2 seconds"
    cmds: |
      sleep 2
      echo "i slept for 2 seconds!"
    parallel: true

  sleep-for-3:
    desc: "sleep for 3 seconds"
    cmds: |
      sleep 3
      echo "i slept for 3 seconds!"
    parallel: true

  sleep-all:
    desc: "run all sleep tasks"
    cmds: |
      echo {{.SLEEP_ALL_SUCCESS_MSG}}
    deps:
      - sleep-for-2
      - sleep-for-3
`
}
