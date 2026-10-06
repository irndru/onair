package controlcenter

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestAppsInNames runs on every platform: the fake reads no Info.plist.
func TestAppsInNames(t *testing.T) {
	dir := t.TempDir()
	mkdir := func(path string) string {
		t.Helper()
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		return path
	}
	display := mkdir(filepath.Join(dir, "Display.app"))
	trimmed := mkdir(filepath.Join(dir, "Trimmed.app"))
	bare := mkdir(filepath.Join(dir, "Bare Name.app"))
	nested := mkdir(filepath.Join(dir, "Utilities", "Nested.app"))
	dup := mkdir(filepath.Join(dir, "Utilities", "Dup.app"))
	mkdir(filepath.Join(dir, "Utilities", "Deeper", "Deep.app"))
	if err := os.Symlink(filepath.Join(dir, "Utilities"), filepath.Join(dir, "Link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	f := useFake(t, &fakeBridge{
		bundles: map[string]bundle{
			display: {id: "a.display", displayName: "Shown", name: "Hidden", camera: true},
			trimmed: {id: "a.trimmed", name: "\u200eTrimmed ", camera: true},
			bare:    {id: "a.bare", camera: true},
			nested:  {id: "a.nested", name: "Nested", camera: true},
			dup:     {id: "a.display", name: "Duplicate", camera: true},
		},
		apps: map[string]*fakeApp{"a.nested": {toggled: true}},
	})
	// The symlinked copy of Utilities must not add anything new.
	f.bundles[filepath.Join(dir, "Link", "Nested.app")] = f.bundles[nested]

	apps, err := appsIn([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	want := []App{
		{BundleID: "a.bare", Name: "Bare Name"},
		{BundleID: "a.display", Name: "Shown"},
		{BundleID: "a.nested", Name: "Nested", Toggled: true},
		{BundleID: "a.trimmed", Name: "Trimmed"},
	}
	if !slices.Equal(apps, want) {
		t.Errorf("appsIn = %+v\nwant %+v", apps, want)
	}
}
