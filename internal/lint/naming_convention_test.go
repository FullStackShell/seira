package lint

import (
	"testing"
)

func TestNamingConventionRule_GoodBash(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-naming/good_bash.sh",
		BaseDir:   "../../testdata/lint-naming",
		Mode:      "library",
		Prefix:    []string{"oop"},
		Shell:     "bash",
	})
	if err != nil {
		t.Fatal(err)
	}
	namingDiags := findDiags(diags, "naming-convention")
	if len(namingDiags) != 0 {
		for _, d := range namingDiags {
			t.Logf("unexpected: %s", d)
		}
		t.Errorf("expected no naming diagnostics for good_bash.sh, got %d", len(namingDiags))
	}
}

func TestNamingConventionRule_BadBash(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-naming/bad_bash.sh",
		BaseDir:   "../../testdata/lint-naming",
		Mode:      "library",
		Prefix:    []string{"oop"},
		Shell:     "bash",
	})
	if err != nil {
		t.Fatal(err)
	}
	namingDiags := findDiags(diags, "naming-convention")
	if len(namingDiags) == 0 {
		t.Fatal("expected naming convention warnings for bad_bash.sh, got none")
	}

	// Should have warnings for: _L, _WRONG_NAME, bad_function
	foundVar := false
	foundFunc := false
	for _, d := range namingDiags {
		t.Logf("diag: %s", d)
		if d.Message != "" {
			if contains(d.Message, "_L") || contains(d.Message, "_WRONG_NAME") {
				foundVar = true
			}
			if contains(d.Message, "bad_function") {
				foundFunc = true
			}
		}
	}
	if !foundVar {
		t.Error("expected warning about bad variable names (_L or _WRONG_NAME)")
	}
	if !foundFunc {
		t.Error("expected warning about bad function name (bad_function)")
	}
}

func TestNamingConventionRule_GoodSh(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-naming/good_sh.sh",
		BaseDir:   "../../testdata/lint-naming",
		Mode:      "library",
		Prefix:    []string{"anysock"},
		Shell:     "sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	namingDiags := findDiags(diags, "naming-convention")
	if len(namingDiags) != 0 {
		for _, d := range namingDiags {
			t.Logf("unexpected: %s", d)
		}
		t.Errorf("expected no naming diagnostics for good_sh.sh, got %d", len(namingDiags))
	}
}

func TestNamingConventionRule_BadSh(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-naming/bad_sh.sh",
		BaseDir:   "../../testdata/lint-naming",
		Mode:      "library",
		Prefix:    []string{"anysock"},
		Shell:     "sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	namingDiags := findDiags(diags, "naming-convention")
	if len(namingDiags) == 0 {
		t.Fatal("expected naming convention warnings for bad_sh.sh, got none")
	}

	// Should have warnings for: ANYSOCK_VERSION, _WRONG_VAR, wrong_function
	for _, d := range namingDiags {
		t.Logf("diag: %s", d)
	}
}

func TestNamingConventionRule_OptOut(t *testing.T) {
	// When prefix is empty, rule should produce no diagnostics
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-naming/bad_bash.sh",
		BaseDir:   "../../testdata/lint-naming",
		Mode:      "library",
		Prefix:    []string{},
		Shell:     "bash",
	})
	if err != nil {
		t.Fatal(err)
	}
	namingDiags := findDiags(diags, "naming-convention")
	if len(namingDiags) != 0 {
		t.Errorf("expected no naming diagnostics when prefix is empty, got %d", len(namingDiags))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
