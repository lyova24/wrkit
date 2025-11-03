package src

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// cmdRootLogic - main function for cmdRoot command
func cmdRootLogic(cmd *cobra.Command, args []string) error {
	// If --mode not provided — root command should act like `run`,
	// take task name as first argument and run it.
	if modeFlag {
		// In subcommands mode, rootCmd does nothing by itself —
		// logic provided by subcommands; showing help, if nothing called.
		return cmd.Help()
	}
	// If --mode not provided, looking for at least one argument - task name.
	if len(args) < 1 {
		return cmd.Help()
	}
	taskName := args[0]
	cfg, err := LoadCombinedConfig(cfgFile, noMaster)
	if err != nil {
		return err
	}
	varsMap := parseVars(varsSlice)
	return RunTaskByName(cfg, taskName, dryRun, verbose, varsMap)
}

// cmdRunLogic - main function for cmdRun command
func cmdRunLogic(_ *cobra.Command, args []string) error {
	taskName := args[0]
	cfg, err := LoadCombinedConfig(cfgFile, noMaster)
	if err != nil {
		return err
	}
	varsMap := parseVars(varsSlice)
	return RunTaskByName(cfg, taskName, dryRun, verbose, varsMap)
}

// cmdListLogic - main function for cmdList command
func cmdListLogic(_ *cobra.Command, _ []string) error {
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
}

// cmdShowLogic - main function for cmdShow command
func cmdShowLogic(_ *cobra.Command, args []string) error {
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
}

// cmdInitLogic - main function for cmdInit command
func cmdInitLogic(_ *cobra.Command, _ []string) error {
	if _, err := os.Stat("wrkit.yaml"); err == nil {
		return fmt.Errorf("wrkit.yaml already exists in current directory")
	}
	return os.WriteFile("wrkit.yaml", []byte(wrkitYAMLExample), 0644)
}

// cmdVersionLogic - main function for cmdVersion command
func cmdVersionLogic(_ *cobra.Command, _ []string) {
	fmt.Println("wrkit", version)
}
