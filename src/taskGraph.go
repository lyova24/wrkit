package src

import (
	"errors"
	"fmt"
)

type TaskNode struct {
	Name string
	Cfg  *TaskConfig
}

type TaskGraph struct {
	Nodes map[string]*TaskNode
	// adjacency: node -> deps
	Deps map[string][]string
}

func BuildGraph(cfg *Config) (*TaskGraph, error) {
	g := &TaskGraph{
		Nodes: map[string]*TaskNode{},
		Deps:  map[string][]string{},
	}
	for name, tcfg := range cfg.Tasks {
		g.Nodes[name] = &TaskNode{
			Name: name,
			Cfg:  tcfg,
		}
		if tcfg.Deps == nil {
			g.Deps[name] = []string{}
		} else {
			g.Deps[name] = append([]string(nil), tcfg.Deps...)
		}
	}
	// Validate deps presence
	for name, deps := range g.Deps {
		for _, d := range deps {
			if _, ok := g.Nodes[d]; !ok {
				return nil, fmt.Errorf("task %q depends on unknown task %q", name, d)
			}
		}
	}
	// Check cycles
	if err := checkCycles(g); err != nil {
		return nil, err
	}
	return g, nil
}

func checkCycles(g *TaskGraph) error {
	// DFS coloring
	const (
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var stack []string
	var visit func(string) error
	visit = func(u string) error {
		color[u] = gray
		stack = append(stack, u)
		for _, v := range g.Deps[u] {
			if color[v] == 0 {
				if err := visit(v); err != nil {
					return err
				}
			} else if color[v] == gray {
				// found cycle
				// build cycle description
				cycle := append([]string{}, stack...)
				cycle = append(cycle, v)
				return errors.New("cycle detected: " + fmt.Sprint(cycle))
			}
		}
		color[u] = black
		stack = stack[:len(stack)-1]
		return nil
	}
	for name := range g.Nodes {
		if color[name] == 0 {
			if err := visit(name); err != nil {
				return err
			}
		}
	}
	return nil
}

// CollectSubgraph returns all tasks needed for the named root (including root).
func (g *TaskGraph) CollectSubgraph(root string) ([]string, error) {
	if _, ok := g.Nodes[root]; !ok {
		return nil, fmt.Errorf("task %q not found", root)
	}
	// order is post-order: dependencies first, root last
	return g.dfsCollect(root), nil
}

// dfsCollect - performs a dfs to collect task dependencies in post-order
func (g *TaskGraph) dfsCollect(root string) []string {
	type frame struct {
		node     string
		expanded bool // false = first visit, true = after children
	}

	stack := []frame{{node: root}}
	visited := map[string]bool{}
	var order []string

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[top.node] && !top.expanded {
			continue
		}
		if top.expanded {
			order = append(order, top.node)
			continue
		}
		visited[top.node] = true
		stack = append(stack, frame{node: top.node, expanded: true})
		for i := len(g.Deps[top.node]) - 1; i >= 0; i-- {
			d := g.Deps[top.node][i]
			if !visited[d] {
				stack = append(stack, frame{node: d})
			}
		}
	}

	return order
}

func (g *TaskGraph) WavesFor(root string) ([][]string, error) {
	order, err := g.CollectSubgraph(root)
	if err != nil {
		return nil, err
	}

	indeg := map[string]int{}
	for k := range g.Nodes {
		indeg[k] = len(g.Deps[k])
	}

	var waves [][]string
	subset := map[string]bool{}
	for _, n := range order {
		subset[n] = true
	}

	done := map[string]bool{}
	for {
		var ready []string
		for n := range subset {
			if done[n] {
				continue
			}
			allDone := true
			for _, dep := range g.Deps[n] {
				if subset[dep] && !done[dep] {
					allDone = false
					break
				}
			}
			if allDone {
				ready = append(ready, n)
			}
		}
		if len(ready) == 0 {
			break
		}
		waves = append(waves, ready)
		for _, r := range ready {
			done[r] = true
		}
	}
	return waves, nil
}
