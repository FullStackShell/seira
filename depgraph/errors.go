package depgraph

import (
	"fmt"
	"strings"
)

// CycleError is returned when circular dependencies are detected.
type CycleError struct {
	Cycles [][]string
}

func (e *CycleError) Error() string {
	var sb strings.Builder
	sb.WriteString("circular dependency detected:\n")
	for _, cycle := range e.Cycles {
		sb.WriteString("  ")
		sb.WriteString(fmt.Sprintf("%s", strings.Join(cycle, " → ")))
		sb.WriteString("\n")
	}
	return sb.String()
}
