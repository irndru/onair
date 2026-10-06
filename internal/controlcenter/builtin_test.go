package controlcenter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveImage(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.png")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	// TempDir may sit behind a symlink, and Abs uses the resolved working directory.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ in, want string }{
		{file, file},
		{"a.png", filepath.Join(wd, "a.png")},
		{filepath.Join(dir, "missing.png"), ""},
		{dir, ""},
		{"", ""},
	}
	for _, tt := range tests {
		got, err := ResolveImage(tt.in)
		if got != tt.want || (err != nil) != (tt.want == "") {
			t.Errorf("ResolveImage(%q) = %q, %v; want %q", tt.in, got, err, tt.want)
		}
	}
}

func TestBuiltin(t *testing.T) {
	if _, err := os.Stat(gradientDir); err != nil {
		t.Skip("no built-in backgrounds:", err)
	}
	for _, name := range Builtin() {
		path, ok := BuiltinPath(name)
		if !ok {
			t.Fatalf("BuiltinPath(%q) not found", name)
		}
		if got, err := ResolveImage(name); got != path || err != nil {
			t.Errorf("ResolveImage(%q) = %q, %v; want %q", name, got, err, path)
		}
	}
}

func TestResolveImageBuiltinFirst(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "blue"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// ./blue means the file, whatever the built-ins.
	if got, err := ResolveImage("./blue"); got != filepath.Join(wd, "blue") || err != nil {
		t.Errorf("ResolveImage(./blue) = %q, %v", got, err)
	}
	if _, err := os.Stat(gradientDir); err != nil {
		t.Skip("no built-in backgrounds:", err)
	}
	want, _ := BuiltinPath("blue")
	for _, name := range []string{"blue", "BLUE"} {
		if got, err := ResolveImage(name); got != want || err != nil {
			t.Errorf("ResolveImage(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
}

func TestBuiltinPathMiss(t *testing.T) {
	if got, ok := BuiltinPath("nope"); ok || got != "" {
		t.Errorf("BuiltinPath(nope) = %q, %v", got, ok)
	}
}
