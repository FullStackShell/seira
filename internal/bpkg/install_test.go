package bpkg

import "testing"

func TestParsePackageRef(t *testing.T) {
	tests := []struct {
		input   string
		user    string
		name    string
		version string
	}{
		{"bpkg/term", "bpkg", "term", "master"},
		{"jwerle/suggest.sh@0.0.1", "jwerle", "suggest.sh", "0.0.1"},
		{"user/name@v1.2.3", "user", "name", "v1.2.3"},
		{"org/pkg@main", "org", "pkg", "main"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, err := ParsePackageRef(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if ref.User != tt.user {
				t.Errorf("user: got %q, want %q", ref.User, tt.user)
			}
			if ref.Name != tt.name {
				t.Errorf("name: got %q, want %q", ref.Name, tt.name)
			}
			if ref.Version != tt.version {
				t.Errorf("version: got %q, want %q", ref.Version, tt.version)
			}
		})
	}
}

func TestParsePackageRef_Invalid(t *testing.T) {
	invalids := []string{
		"nouser",
		"/name",
		"user/",
		"",
	}
	for _, s := range invalids {
		t.Run(s, func(t *testing.T) {
			_, err := ParsePackageRef(s)
			if err == nil {
				t.Errorf("expected error for %q", s)
			}
		})
	}
}

func TestPackageRef_String(t *testing.T) {
	ref := &PackageRef{User: "bpkg", Name: "term", Version: "0.1.0"}
	if got := ref.String(); got != "bpkg/term@0.1.0" {
		t.Errorf("got %q, want %q", got, "bpkg/term@0.1.0")
	}

	ref2 := &PackageRef{User: "bpkg", Name: "term", Version: "master"}
	if got := ref2.String(); got != "bpkg/term" {
		t.Errorf("got %q, want %q", got, "bpkg/term")
	}
}
