package src

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	concurrency int
	dryRun      bool
	verbose     bool
	varsSlice   []string
	version     = "0.1.0"
	noMaster    bool
)

func Execute() {
	root := &cobra.Command{
		Use:   "wrkit",
		Short: "wrkit — YAML-powered tiny make-like runner",
		Long:  "wrkit — a small, fast task runner driven by YAML files.",
	}

	root.PersistentFlags().StringVarP(&cfgFile, "file", "f", "wrkit.yaml", "wrkit YAML configuration file")
	root.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 4, "Number of tasks to run concurrently")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print what would be done without executing")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	root.PersistentFlags().StringArrayVarP(&varsSlice, "var", "V", []string{}, "Variables to pass to templates (key=value). Can be repeated.")
	root.PersistentFlags().BoolVar(&noMaster, "no-master", false, "Ignore global ~/.wrkit.master.yaml and use only local wrkit.yaml")

	root.AddCommand(cmdRun())
	root.AddCommand(cmdList())
	root.AddCommand(cmdShow())
	root.AddCommand(cmdInit())
	root.AddCommand(cmdVersion())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run task and its dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			taskName := args[0]
			cfg, err := LoadCombinedConfig(cfgFile, noMaster)
			if err != nil {
				return err
			}
			varsMap := parseVars(varsSlice)
			return RunTaskByName(cfg, taskName, concurrency, dryRun, verbose, varsMap)
		},
	}
	return cmd
}

func cmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks in the config",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := LoadCombinedConfig(cfgFile, noMaster)
			if err != nil {
				return err
			}
			for name, t := range cfg.Tasks {
				desc := t.Desc
				if desc == "" {
					desc = "-"
				}
				fmt.Printf("%-20s %s\n", name, desc)
			}
			return nil
		},
	}
	return cmd
}

func cmdShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [task]",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := LoadCombinedConfig(cfgFile, noMaster)
			if err != nil {
				return err
			}
			t, ok := cfg.Tasks[name]
			if !ok {
				return fmt.Errorf("task %q not found", name)
			}
			fmt.Printf("name: %s\n", name)
			fmt.Printf("desc: %s\n", t.Desc)
			fmt.Printf("dir:  %s\n", t.Dir)
			if len(t.Deps) > 0 {
				fmt.Printf("deps: %s\n", strings.Join(t.Deps, ", "))
			}
			if len(t.Cmds) > 0 {
				fmt.Println("cmds:")
				for _, c := range t.Cmds {
					fmt.Printf("  - %s\n", c)
				}
			}
			if len(t.Env) > 0 {
				fmt.Println("env:")
				for k, v := range t.Env {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}
			fmt.Printf("parallel: %v\n", t.Parallel)
			return nil
		},
	}
	return cmd
}

func cmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create example wrkit.yaml in current directory",
		RunE: func(c *cobra.Command, args []string) error {
			if _, err := os.Stat("wrkit.yaml"); err == nil {
				return fmt.Errorf("wrkit.yaml already exists in current directory")
			}
			content := sampleYAML()
			return os.WriteFile("wrkit.yaml", []byte(content), 0644)
		},
	}
	return cmd
}

func cmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("wrkit", version)
		},
	}
	return cmd
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
