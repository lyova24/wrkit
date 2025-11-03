package src

import (
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
	modeFlag    bool
)

func cmdRoot() *cobra.Command {
	return &cobra.Command{Use: "wrkit [flags] [task-name]",
		Short: "wrkit — YAML-powered tiny make-like runner",
		Long:  cmdRootLongDescription,
		RunE:  cmdRootLogic,
	}
}

func cmdRun() *cobra.Command {
	return &cobra.Command{
		Use:   "run [task]",
		Short: "Run task and its dependencies",
		Args:  cobra.ExactArgs(1),
		RunE:  cmdRunLogic,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return getTaskNameCompletions(toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tasks in the config",
		RunE:  cmdListLogic,
	}
}

func cmdShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show [task]",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE:  cmdShowLogic,
	}
}

func cmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create example wrkit.yaml in current directory",
		RunE:  cmdInitLogic,
	}
}

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   cmdVersionLogic,
	}
}
