package src

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

func renderTemplate(tmpl string, vars map[string]string) (string, error) {
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

func parseVars(slice []string) map[string]string {
	out := map[string]string{}
	for _, s := range slice {
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 1 {
			out[parts[0]] = ""
		} else {
			out[parts[0]] = parts[1]
		}
	}
	return out
}

func getTaskNameCompletions(toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := LoadCombinedConfig(cfgFile, noMaster)
	if err != nil || cfg == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for name := range cfg.Tasks {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
