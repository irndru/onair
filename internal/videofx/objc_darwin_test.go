package videofx

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestLoad fails when Apple renames or removes the private symbols the
// background needs. It only logs the optional ones, which older macOS lacks.
func TestLoad(t *testing.T) {
	b, err := newBridge()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range Effects {
		if err := b.effectErr(e); err != nil {
			t.Logf("%s: %v", e, err)
		}
	}
	if err := b.micErr(); err != nil {
		t.Logf("mic: %v", err)
	}
}

func TestStringRoundTrip(t *testing.T) {
	if _, err := openBridge(); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"", "com.apple.PhotoBooth", "/tmp/with space.png", "café 背景 🎥"} {
		if got := goString(nsString(s)); got != s {
			t.Errorf("round trip of %q = %q", s, got)
		}
	}
	if got := goString(0); got != "" {
		t.Errorf("goString(nil) = %q", got)
	}
}

// writeApp creates dir/<file>.app with a hand-written Info.plist.
func writeApp(t *testing.T, dir, file, id, name string, camera bool) string {
	t.Helper()
	app := filepath.Join(dir, file+".app")
	if err := os.MkdirAll(filepath.Join(app, "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	usage := ""
	if camera {
		usage = "<key>NSCameraUsageDescription</key><string>Video calls.</string>"
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key><string>%s</string>
	<key>CFBundleName</key><string>%s</string>
	%s
</dict>
</plist>
`, id, name, usage)
	if err := os.WriteFile(filepath.Join(app, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
	return app
}

func TestBundleInfo(t *testing.T) {
	b, err := openBridge()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	got, ok := b.bundleInfo(writeApp(t, dir, "Caller", "test.videobg.caller", "Caller", true))
	want := bundle{id: "test.videobg.caller", name: "Caller", camera: true}
	if !ok || got != want {
		t.Errorf("bundleInfo = %+v, %v; want %+v", got, ok, want)
	}
	if got, ok := b.bundleInfo(filepath.Join(dir, "Missing.app")); ok {
		t.Errorf("bundleInfo of a missing app = %+v", got)
	}
}

func TestAppsIn(t *testing.T) {
	dir := t.TempDir()
	writeApp(t, dir, "Caller", "test.videobg.caller", "Caller", true)
	writeApp(t, dir, "Chrome", "com.google.Chrome", "Chrome", false)
	writeApp(t, dir, "Editor", "test.videobg.editor", "Editor", false)
	writeApp(t, filepath.Join(dir, "Utilities"), "Nested", "test.videobg.nested", "", true)
	writeApp(t, dir, "Copy", "test.videobg.caller", "Copy", true)

	apps, err := appsIn([]string{dir, filepath.Join(dir, "missing")})
	if err != nil {
		t.Fatal(err)
	}
	// Toggled depends on the machine's camera history for Chrome.
	for i := range apps {
		apps[i].Toggled = false
	}
	want := []App{
		{BundleID: "com.google.Chrome", Name: "Chrome", Known: true},
		{BundleID: "test.videobg.caller", Name: "Caller"},
		{BundleID: "test.videobg.nested", Name: "Nested"},
	}
	if !slices.Equal(apps, want) {
		t.Errorf("appsIn = %+v\nwant %+v", apps, want)
	}
}
