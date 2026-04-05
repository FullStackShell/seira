package depgraph

import "fmt"

// TopologicalSort returns file paths in dependency order (dependencies first).
// Returns an error if a cycle is detected.
func (g *Graph) TopologicalSort() ([]string, error) {
	cycles := g.DetectCycles()
	if len(cycles) > 0 {
		return nil, fmt.Errorf("circular dependency detected: %v", cycles[0])
	}

	visited := make(map[string]bool)
	var order []string

	var visit func(path string)
	visit = func(path string) {
		if visited[path] {
			return
		}
		visited[path] = true
		for _, dep := range g.edges[path] {
			visit(dep)
		}
		order = append(order, path)
	}

	for path := range g.nodes {
		visit(path)
	}

	return order, nil
}

// DetectCycles returns all cycles found in the graph.
// Each cycle is represented as a slice of paths forming the cycle.
func (g *Graph) DetectCycles() [][]string {
	const (
		white = iota // unvisited
		gray         // in current DFS path
		black        // fully processed
	)

	color := make(map[string]int)
	parent := make(map[string]string)
	var cycles [][]string

	var dfs func(node string)
	dfs = func(node string) {
		color[node] = gray
		for _, next := range g.edges[node] {
			if color[next] == gray {
				// Found a cycle — reconstruct it
				cycle := []string{next}
				cur := node
				for cur != next {
					cycle = append([]string{cur}, cycle...)
					cur = parent[cur]
				}
				cycle = append([]string{next}, cycle...)
				cycles = append(cycles, cycle)
			} else if color[next] == white {
				parent[next] = node
				dfs(next)
			}
		}
		color[node] = black
	}

	for path := range g.nodes {
		if color[path] == white {
			dfs(path)
		}
	}

	return cycles
}
