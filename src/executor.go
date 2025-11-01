package src

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type execResult struct {
	name string
	err  error
}

func RunTaskByName(cfg *Config, name string, concurrency int, dryRun bool, verbose bool, vars map[string]string) error {
	g, err := BuildGraph(cfg)
	if err != nil {
		return err
	}
	order, err := g.CollectSubgraph(name)
	if err != nil {
		return err
	}

	// Build map of in-degrees for tasks within subgraph
	sub := map[string]bool{}
	for _, n := range order {
		sub[n] = true
	}
	indeg := map[string]int{}
	dependents := map[string][]string{} // u -> list of tasks that depend on u
	for u := range sub {
		indeg[u] = 0
	}
	for u := range sub {
		for _, d := range g.Deps[u] {
			if !sub[d] {
				continue
			}
			indeg[u]++
			dependents[d] = append(dependents[d], u)
		}
	}

	// Merge vars: config vars -> env -> CLI vars
	mergedVars := map[string]string{}
	for k, v := range cfg.Vars {
		mergedVars[k] = v
	}
	// environment variables available as env.VAR
	envMap := map[string]string{}
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	// expose env.* to templates by prefix "env."
	// also allow direct env variable override
	for k, v := range envMap {
		mergedVars["env."+k] = v
	}
	for k, v := range vars {
		mergedVars[k] = v
	}

	// Channel of ready tasks
	ready := make(chan string, len(sub))
	for u, d := range indeg {
		if d == 0 {
			ready <- u
		}
	}
	// Worker pool
	if concurrency <= 0 {
		concurrency = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan execResult, len(sub))
	_ = map[string]bool{}
	var _ sync.Mutex

	// function to pick and run a task
	runTaskFunc := func(taskName string) {
		defer wg.Done()
		tnode := g.Nodes[taskName]
		if dryRun {
			fmt.Printf("[dry-run] task %s\n", taskName)
			results <- execResult{name: taskName, err: nil}
			return
		}
		// execute task commands
		if verbose {
			fmt.Printf("[start] task %s\n", taskName)
		} else {
			fmt.Printf("→ %s\n", taskName)
		}
		err := executeTaskCommands(ctx, tnode, mergedVars, verbose)
		if err != nil {
			results <- execResult{name: taskName, err: err}
			return
		}
		if verbose {
			fmt.Printf("[done]  task %s\n", taskName)
		}
		results <- execResult{name: taskName, err: nil}
	}

	// worker goroutines
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-workerCtx.Done():
					return
				case tname, ok := <-ready:
					if !ok {
						return
					}
					// If this task is marked not parallel, we should run it alone
					tnode := g.Nodes[tname]
					if tnode.Cfg.Parallel {
						// run in separate goroutine to allow parallelism
						wg.Add(1)
						go runTaskFunc(tname)
					} else {
						// run here (this worker) to serialize non-parallel tasks
						if dryRun {
							fmt.Printf("[dry-run] task %s\n", tname)
							results <- execResult{name: tname, err: nil}
						} else {
							if verbose {
								fmt.Printf("[start] task %s\n", tname)
							} else {
								fmt.Printf("→ %s\n", tname)
							}
							err := executeTaskCommands(ctx, tnode, mergedVars, verbose)
							if err != nil {
								results <- execResult{name: tname, err: err}
								// cancel everything
								workerCancel()
								return
							}
							if verbose {
								fmt.Printf("[done]  task %s\n", tname)
							}
							results <- execResult{name: tname, err: nil}
						}
					}
				}
			}
		}()
	}

	// Coordinator: collect results and push new ready nodes
	doneCount := 0
	total := len(sub)
	// track completed
	completed := map[string]bool{}
	for doneCount < total {
		select {
		case res := <-results:
			if res.err != nil {
				// cancel workers and return error
				cancel()
				close(ready)
				return fmt.Errorf("task %s failed: %w", res.name, res.err)
			}
			// mark completed
			completed[res.name] = true
			doneCount++
			// decrease indeg of dependents
			for _, dep := range dependents[res.name] {
				indeg[dep]--
				if indeg[dep] == 0 {
					// all deps done -> schedule
					select {
					case ready <- dep:
					default:
						// should not happen, but in case channel full, spawn goroutine to send
						go func(n string) { ready <- n }(dep)
					}
				}
			}
		case <-ctx.Done():
			close(ready)
			return errors.New("execution canceled")
		}
	}

	// all done
	close(ready)
	// tiny wait for goroutines to finish
	// we used WaitGroup but several goroutines incremented; to simplify, sleep shortly
	time.Sleep(50 * time.Millisecond)
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
