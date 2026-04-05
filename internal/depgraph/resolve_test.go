package depgraph

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/Hayao0819/seira/internal/shellparse"
)

func TestResolve_Simple(t *testing.T) {
	parser := shellparse.NewParser()
	entry, _ := filepath.Abs("../../testdata/simple/main.sh")

	g, err := Resolve(parser, entry, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !g.HasNode(entry) {
		t.Errorf("graph should contain entrypoint %s", entry)
	}
	if len(g.Flatten()) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.Flatten()))
	}
}

func TestResolve_WithDeps(t *testing.T) {
	parser := shellparse.NewParser()
	entry, _ := filepath.Abs("../../testdata/deps/main.sh")

	g, err := Resolve(parser, entry, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(g.Flatten()) != 2 {
		t.Errorf("expected 2 nodes, got %d: %v", len(g.Flatten()), g.Flatten())
	}

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatal(err)
	}

	// helper.sh should come before main.sh in topological order
	helperPath, _ := filepath.Abs("../../testdata/deps/lib/helper.sh")
	idx := make(map[string]int)
	for i, p := range order {
		idx[p] = i
	}
	if idx[helperPath] > idx[entry] {
		t.Errorf("helper should come before main in topo order, got: %v", order)
	}
}

func TestResolve_Circular(t *testing.T) {
	parser := shellparse.NewParser()
	entry, _ := filepath.Abs("../../testdata/circular/a.sh")

	_, err := Resolve(parser, entry, nil)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}

	var cycleErr *CycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected CycleError, got %T: %v", err, err)
	}
}

func TestResolve_VarPath(t *testing.T) {
	parser := shellparse.NewParser()
	entry, _ := filepath.Abs("../../testdata/varpath/main.sh")

	g, err := Resolve(parser, entry, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(g.Flatten()) != 2 {
		t.Errorf("expected 2 nodes, got %d: %v", len(g.Flatten()), g.Flatten())
	}
}
