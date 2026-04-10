package doc

// FileDoc represents documentation extracted from a single shell script file.
type FileDoc struct {
	Path        string     // file path (relative to base dir)
	Name        string     // @file tag value
	Brief       string     // @brief tag value
	Description string     // @description tag value (file-level)
	Sections    []Section  // grouped functions under @section headers
	Functions   []FuncDoc  // functions not belonging to any section
	Variables   []Variable // @set tags at file level
}

// Section groups functions under a named heading.
type Section struct {
	Name      string
	Functions []FuncDoc
}

// FuncDoc represents documentation for a single function.
type FuncDoc struct {
	Name        string
	Description string
	Params      []Param
	Options     []Option
	ExitCodes   []ExitCode
	Stdin       string
	Stdout      string
	Returns     string
	Examples    []string
	SeeAlso     []string
	Internal    bool
	Set         []Variable
	Line        int // source line number
}

// Param represents a function parameter documented with @arg.
type Param struct {
	Name        string // e.g. "$1", "$@"
	Description string
}

// Option represents a flag documented with @option.
type Option struct {
	Flags       string // e.g. "-f --force"
	Description string
}

// ExitCode documents exit codes via @exitcode.
type ExitCode struct {
	Code        string
	Description string
}

// Variable documents a global variable via @set.
type Variable struct {
	Name        string
	Description string
}

// ProjectDoc is the top-level documentation for an entire project (multi-file).
type ProjectDoc struct {
	Title string
	Files []FileDoc
}
