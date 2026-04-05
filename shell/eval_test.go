package shell_test

import (
	"testing"

	"github.com/Hayao0819/seira/shell"
)

func TestEvalSh(t *testing.T) {
	cases := []struct {
		cmd    string
		stdout string
		stderr string
		env    map[string]string
	}{
		{"echo Hello, World!", "Hello, World!\n", "", nil},
		{"echo Hello, World! >&2", "", "Hello, World!\n", nil},
		{"echo $FOO", "bar\n", "", map[string]string{"FOO": "bar"}},
	}
	for _, c := range cases {
		stdout, stderr, _, err := shell.EvalSh(c.cmd, c.env)
		if stdout != c.stdout {
			t.Errorf("stdout: got %q, want %q", stdout, c.stdout)
		}
		if stderr != c.stderr {
			t.Errorf("stderr: got %q, want %q", stderr, c.stderr)
		}
		if err != nil {
			t.Errorf("err: got %v, want nil", err)
		}
	}
}
