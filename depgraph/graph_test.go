package depgraph

import (
	"testing"
)

func TestTopologicalSort_Linear(t *testing.T) {
	g := New()
	g.AddNode("a", nil)
	g.AddNode("b", nil)
	g.AddNode("c", nil)
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatal(err)
	}

	// c must come before b, b before a
	idx := make(map[string]int)
	for i, p := range order {
		idx[p] = i
	}
	if idx["c"] > idx["b"] {
		t.Errorf("c should come before b, got order: %v", order)
	}
	if idx["b"] > idx["a"] {
		t.Errorf("b should come before a, got order: %v", order)
	}
}

func TestTopologicalSort_Diamond(t *testing.T) {
	// a → {b, c} → d
	g := New()
	g.AddNode("a", nil)
	g.AddNode("b", nil)
	g.AddNode("c", nil)
	g.AddNode("d", nil)
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatal(err)
	}

	idx := make(map[string]int)
	for i, p := range order {
		idx[p] = i
	}
	if idx["d"] > idx["b"] || idx["d"] > idx["c"] {
		t.Errorf("d should come before b and c, got order: %v", order)
	}
	if idx["b"] > idx["a"] || idx["c"] > idx["a"] {
		t.Errorf("b and c should come before a, got order: %v", order)
	}
}

func TestDetectCycles(t *testing.T) {
	g := New()
	g.AddNode("a", nil)
	g.AddNode("b", nil)
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")

	cycles := g.DetectCycles()
	if len(cycles) == 0 {
		t.Error("expected at least one cycle, got none")
	}
}

func TestTopologicalSort_CycleError(t *testing.T) {
	g := New()
	g.AddNode("a", nil)
	g.AddNode("b", nil)
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")

	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("expected error for cycle, got nil")
	}
}

func TestNoCycle(t *testing.T) {
	g := New()
	g.AddNode("a", nil)
	g.AddNode("b", nil)
	g.AddEdge("a", "b")

	cycles := g.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}
