package lint

import (
	"testing"
)

func findDiags(diags []Diagnostic, rule string) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Rule == rule {
			out = append(out, d)
		}
	}
	return out
}

func TestBashSourceRule_Concat(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-bashsource/main.sh",
		BaseDir:   "../../testdata/lint-bashsource",
		Mode:      "concat",
	})
	if err != nil {
		t.Fatal(err)
	}

	bashDiags := findDiags(diags, "bash-source")
	if len(bashDiags) == 0 {
		t.Fatal("expected BASH_SOURCE warnings, got none")
	}

	files := map[string]bool{}
	for _, d := range bashDiags {
		files[d.File] = true
		t.Logf("diag: %s", d)
	}
	if !files["main.sh"] {
		t.Error("expected warning in main.sh")
	}
	if !files["lib/helper.sh"] {
		t.Error("expected warning in lib/helper.sh")
	}
}

func TestBashSourceRule_Tarball(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-bashsource/main.sh",
		BaseDir:   "../../testdata/lint-bashsource",
		Mode:      "tarball",
	})
	if err != nil {
		t.Fatal(err)
	}
	bashDiags := findDiags(diags, "bash-source")
	if len(bashDiags) != 0 {
		t.Errorf("expected no bash-source warnings for tarball mode, got %d", len(bashDiags))
	}
}

func TestConditionalSourceRule(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-condsource/main.sh",
		BaseDir:   "../../testdata/lint-condsource",
		Mode:      "concat",
	})
	if err != nil {
		t.Fatal(err)
	}

	condDiags := findDiags(diags, "conditional-source")
	if len(condDiags) == 0 {
		t.Fatal("expected conditional source warnings, got none")
	}
	for _, d := range condDiags {
		t.Logf("diag: %s", d)
	}
}

func TestConditionalSourceRule_Tarball(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-condsource/main.sh",
		BaseDir:   "../../testdata/lint-condsource",
		Mode:      "tarball",
	})
	if err != nil {
		t.Fatal(err)
	}
	condDiags := findDiags(diags, "conditional-source")
	if len(condDiags) != 0 {
		t.Errorf("expected no warnings for tarball mode, got %d", len(condDiags))
	}
}

func TestFuncCollisionRule(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/lint-collision/main.sh",
		BaseDir:   "../../testdata/lint-collision",
		Mode:      "concat",
	})
	if err != nil {
		t.Fatal(err)
	}

	collisionDiags := findDiags(diags, "func-collision")
	if len(collisionDiags) == 0 {
		t.Fatal("expected function collision warnings, got none")
	}
	for _, d := range collisionDiags {
		t.Logf("diag: %s", d)
	}
}

func TestLint_Clean(t *testing.T) {
	diags, err := Lint(Config{
		InputPath: "../../testdata/simple/main.sh",
		BaseDir:   "../../testdata/simple",
		Mode:      "concat",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		for _, d := range diags {
			t.Logf("unexpected: %s", d)
		}
		t.Errorf("expected no diagnostics for clean project, got %d", len(diags))
	}
}
