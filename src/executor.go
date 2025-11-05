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

	subgraph, err := g.CollectSubgraph(name)
	if err != nil {
		return err
	}

	mergedVars := MergeVars(cfg, vars)

	// Определяем тип каждой задачи: deps-task или main-task
	taskType := make(map[string]string)
	for _, t := range subgraph {
		taskType[t] = "deps-task"
	}
	if len(subgraph) > 0 {
		taskType[name] = "main-task"
	}

	// Fetching dependency waves
	waves, err := g.WavesFor(name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Для отслеживания успешности выполнения задач
	taskResults := make(map[string]error)

	for waveIdx, wave := range waves {
		if verbose {
			fmt.Printf("\n[wave %d] %v\n", waveIdx+1, wave)
		}

		var parallelBatch []*TaskNode
		var wg sync.WaitGroup
		errCh := make(chan error, len(wave))
		resultMu := sync.Mutex{}

		runParallelBatch := func() error {
			if len(parallelBatch) == 0 {
				return nil
			}
			if verbose {
				var names []string
				for _, n := range parallelBatch {
					names = append(names, n.Name)
				}
				fmt.Printf("→ running parallel group: %v\n", names)
			}
			for _, node := range parallelBatch {
				wg.Add(1)
				go func(n *TaskNode) {
					defer wg.Done()
					tType := taskType[n.Name]
					if tType == "" {
						tType = "deps-task"
					}
					if dryRun {
						fmt.Printf("[dry-run][%s] task %s\n", tType, n.Name)
						resultMu.Lock()
						taskResults[n.Name] = nil
						resultMu.Unlock()
						return
					}
					if verbose {
						fmt.Printf("→ [%s] (par) %s\n", tType, n.Name)
					} else {
						fmt.Printf("→ [%s] %s\n", tType, n.Name)
					}
					err := executeTaskCommands(ctx, n, mergedVars, verbose, tType)
					resultMu.Lock()
					taskResults[n.Name] = err
					resultMu.Unlock()
					if err != nil {
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

		for _, taskName := range wave {
			node := g.Nodes[taskName]
			tType := taskType[node.Name]
			if tType == "" {
				tType = "deps-task"
			}

			if node.Cfg.Parallel {
				parallelBatch = append(parallelBatch, node)
				continue
			}

			if err := runParallelBatch(); err != nil {
				return err
			}
			if verbose {
				fmt.Printf("→ [%s] (seq) %s\n", tType, node.Name)
			} else {
				fmt.Printf("→ [%s] %s\n", tType, node.Name)
			}
			if dryRun {
				fmt.Printf("[dry-run][%s] task %s\n", tType, node.Name)
				taskResults[node.Name] = nil
				continue
			}
			err := executeTaskCommands(ctx, node, mergedVars, verbose, tType)
			taskResults[node.Name] = err
			if err != nil {
				return fmt.Errorf("task %s failed: %w", node.Name, err)
			}
		}

		if err := runParallelBatch(); err != nil {
			return err
		}
	}

	// После выполнения основной задачи — запустить post-tasks
	if mainNode, ok := g.Nodes[name]; ok && len(mainNode.Cfg.Post) > 0 {
		mainTaskErr := taskResults[name]
		for _, post := range mainNode.Cfg.Post {
			shouldRun := false
			whenType := normalizeWhen(post.When)
			switch whenType {
			case "success":
				shouldRun = mainTaskErr == nil
			case "always":
				shouldRun = true
			case "fail":
				shouldRun = mainTaskErr != nil
			default:
				_, err := fmt.Fprintf(os.Stderr, "Unknown 'when' value for post-task %q: %q (skipped)\n", post.Name, post.When)
				if err != nil {
					return err
				}
				continue
			}
			if !shouldRun {
				if verbose {
					fmt.Printf("[post-task:%s] skipping %s (when=%s)\n", whenType, post.Name, post.When)
				}
				continue
			}
			postNode, ok := g.Nodes[post.Name]
			if !ok {
				_, err := fmt.Fprintf(os.Stderr, "Post-task %q not found (skipped)\n", post.Name)
				if err != nil {
					return err
				}
				continue
			}
			logPrefix := fmt.Sprintf("[post-task:%s]", whenType)
			if verbose {
				fmt.Printf("→ %s running %s (when=%s)\n", logPrefix, post.Name, post.When)
			} else {
				fmt.Printf("→ %s %s\n", logPrefix, post.Name)
			}
			err := executeTaskCommands(ctx, postNode, mergedVars, verbose, fmt.Sprintf("post-task:%s", whenType))
			if err != nil {
				_, err := fmt.Fprintf(os.Stderr, "Post-task %q failed: %v\n", post.Name, err)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// normalizeWhen приводит when к одному из: success, fail, always
func normalizeWhen(when string) string {
	switch when {
	case "", "success":
		return "success"
	case "always":
		return "always"
	case "fails", "failed", "fail":
		return "fail"
	default:
		return when
	}
}

func executeTaskCommands(ctx context.Context, node *TaskNode, vars map[string]string, verbose bool, taskType string) error {
	t := node.Cfg
	for _, rawCmd := range t.Cmds {
		cmdStr, err := renderTemplate(rawCmd, vars)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("[cmd][%s] %s\n", taskType, cmdStr)
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
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", cmdStr, err)
		}
	}
	return nil
}
