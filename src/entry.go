package src

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func Entrypoint() {
	// Check os.Args — if subcommands mode required (--mode / -m).
	// Need this, to register subcommands only then user provided --mode flag.
	modeFlag = false
	for _, a := range os.Args[1:] {
		if a == "-m" || a == "--mode" || strings.HasPrefix(a, "-m=") || strings.HasPrefix(a, "--mode=") {
			modeFlag = true
			break
		}
	}

	cmdRoot := cmdRoot()
	cmdRoot.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if modeFlag {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 0 {
			return getTaskNameCompletions(toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Common persistent-flags (will be available in subcommands and base modes both).
	cmdRoot.PersistentFlags().StringVarP(&cfgFile, "file", "f", "wrkit.yaml", "wrkit YAML configuration file")
	cmdRoot.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 4, "Number of tasks to run concurrently")
	cmdRoot.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print what would be done without executing")
	cmdRoot.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmdRoot.PersistentFlags().StringArrayVarP(&varsSlice, "var", "V", []string{}, "Variables to pass to templates (key=value). Can be repeated.")
	cmdRoot.PersistentFlags().BoolVar(&noMaster, "no-master", false, "Ignore global ~/.wrkit.master.yaml and use only local wrkit.yaml")

	// Registering flag --mode / -m; default value — result of os.Args check.
	cmdRoot.PersistentFlags().BoolVarP(&modeFlag, "mode", "m", modeFlag,
		"Enable subcommand mode. When set, use subcommands (run, list, show, init, version).\n"+
			"When omitted, the first positional argument is treated as a task name (wrkit <task-name>).")

	// Registering subcommands only when --mode provided
	if modeFlag {
		cmdRoot.AddCommand(cmdRun())
		cmdRoot.AddCommand(cmdList())
		cmdRoot.AddCommand(cmdShow())
		cmdRoot.AddCommand(cmdInit())
		cmdRoot.AddCommand(cmdVersion())
	}

	if err := cmdRoot.Execute(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
