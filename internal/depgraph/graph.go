package depgraph

import (
	"github.com/Hayao0819/seira/internal/shellparse"
)

// Node represents a script file in the dependency graph.
type Node struct {
	Path   string
	Script *shellparse.Script
}

// Graph is a directed acyclic graph of script dependencies.
type Graph struct {
	nodes map[string]*Node
	edges map[string][]string // from → []to (dependencies)
}

func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
	}
}

func (g *Graph) AddNode(path string, s *shellparse.Script) {
	g.nodes[path] = &Node{Path: path, Script: s}
}

func (g *Graph) AddEdge(from, to string) {
	g.edges[from] = append(g.edges[from], to)
}

func (g *Graph) HasNode(path string) bool {
	_, ok := g.nodes[path]
	return ok
}

func (g *Graph) Nodes() map[string]*Node {
	return g.nodes
}

func (g *Graph) Node(path string) *Node {
	return g.nodes[path]
}

func (g *Graph) EdgesFrom(path string) []string {
	return g.edges[path]
}

// Flatten returns all file paths in the graph (unordered).
func (g *Graph) Flatten() []string {
	paths := make([]string, 0, len(g.nodes))
	for p := range g.nodes {
		paths = append(paths, p)
	}
	return paths
}
