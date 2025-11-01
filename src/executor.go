package src

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

func RunTaskByName(cfg *Config, name string, dryRun bool, verbose bool, vars map[string]string) error {
	g, err := BuildGraph(cfg)
	if err != nil {
		return err
	}

	_, err = g.CollectSubgraph(name)
	if err != nil {
		return err
	}

	mergedVars := MergeVars(cfg, vars)

	// Получаем волны зависимостей
	waves, err := g.WavesFor(name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for waveIdx, wave := range waves {
		if verbose {
			fmt.Printf("\n[wave %d] %v\n", waveIdx+1, wave)
		}

		var parallelBatch []*TaskNode
		var wg sync.WaitGroup
		errCh := make(chan error, len(wave))

		runParallelBatch := func() error {
			if len(parallelBatch) == 0 {
				return nil
			}
			if verbose {
				names := []string{}
				for _, n := range parallelBatch {
					names = append(names, n.Name)
				}
				fmt.Printf("→ running parallel group: %v\n", names)
			}
			for _, node := range parallelBatch {
				wg.Add(1)
				go func(n *TaskNode) {
					defer wg.Done()
					if dryRun {
						fmt.Printf("[dry-run] task %s\n", n.Name)
						return
					}
					if verbose {
						fmt.Printf("→ (par) %s\n", n.Name)
					} else {
						fmt.Printf("→ %s\n", n.Name)
					}
					if err := executeTaskCommands(ctx, n, mergedVars, verbose); err != nil {
						errCh <- fmt.Errorf("task %s failed: %w", n.Name, err)
					}
				}(node)
			}
			wg.Wait()
			close(errCh)
			for e := range errCh {
				if e != nil {
					return e
				}
			}
			parallelBatch = nil
			return nil
		}

		for _, tname := range wave {
			node := g.Nodes[tname]

			if node.Cfg.Parallel {
				parallelBatch = append(parallelBatch, node)
				continue
			}

			if err := runParallelBatch(); err != nil {
				return err
			}
			if verbose {
				fmt.Printf("→ (seq) %s\n", node.Name)
			} else {
				fmt.Printf("→ %s\n", node.Name)
			}
			if dryRun {
				fmt.Printf("[dry-run] task %s\n", node.Name)
				continue
			}
			if err := executeTaskCommands(ctx, node, mergedVars, verbose); err != nil {
				return fmt.Errorf("task %s failed: %w", node.Name, err)
			}
		}

		if err := runParallelBatch(); err != nil {
			return err
		}
	}

	fmt.Println("\n✅ all tasks completed successfully")
	return nil
}

func executeTaskCommands(ctx context.Context, node *TaskNode, vars map[string]string, verbose bool) error {
	t := node.Cfg
	for _, rawCmd := range t.Cmds {
		cmdStr, err := renderTemplate(rawCmd, vars)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("[cmd] %s\n", cmdStr)
		}
		// split shell? use `sh -c` to allow pipelines and shell features
		// set up exec.Cmd
		// change dir if specified
		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		if t.Dir != "" {
			cmd.Dir = t.Dir
		} else {
			// default to current working dir
			cmd.Dir, _ = os.Getwd()
		}
		// merge environment: inherited + task env
		env := os.Environ()
		for k, v := range t.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", cmdStr, err)
		}
	}
	return nil
}
