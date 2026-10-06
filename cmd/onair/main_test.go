package main

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/irndru/onair/internal/controlcenter"
)

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		nil, {"bogus"}, {"set", "blue"}, {"on"}, {"studio", "on"}, {"mic", "standard"},
		{"portrait"}, {"portrait", "photo booth"}, {"studio-light", "ON"}, {"edge", "on"},
		{"background"}, {"mic-mode"}, {"mic-mode", "isolation"},
	} {
		err := run(args, io.Discard)
		var usageErr usageError
		if !errors.As(err, &usageErr) {
			t.Errorf("run(%q) = %v, want a usage error", args, err)
		}
	}
}

func TestHelp(t *testing.T) {
	for _, arg := range []string{"help", "-h", "--help"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil {
			t.Fatal(err)
		}
		if out.String() != usage {
			t.Errorf("run(%q) wrote %q, want usage", arg, out.String())
		}
	}
}

func TestVersion(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"version"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "onair ") {
		t.Errorf("version wrote %q", out.String())
	}
}

func TestBackgrounds(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"backgrounds"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, name := range controlcenter.Builtin() {
		if !strings.Contains(out.String(), name+" ") {
			t.Errorf("backgrounds does not list %q", name)
		}
	}
}

func TestBackgroundMissingImage(t *testing.T) {
	err := run([]string{"background", filepath.Join(t.TempDir(), "missing.png")}, io.Discard)
	var usageErr usageError
	if err == nil || errors.As(err, &usageErr) {
		t.Errorf("err = %v, want a non-usage error", err)
	}
}

func TestEdgeLight(t *testing.T) {
	if err := run([]string{"edge-light", "on"}, io.Discard); err != errEdgeLight {
		t.Errorf("err = %v, want errEdgeLight", err)
	}
}

func TestSelectApps(t *testing.T) {
	apps := []controlcenter.App{
		{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth", Known: true},
		{BundleID: "com.example.Notes", Name: "Notes"},
		{BundleID: "us.zoom.xos", Name: "zoom.us", Toggled: true},
	}
	tests := []struct {
		name string
		apps []controlcenter.App
		args []string
		want []string
	}{
		{"defaults", apps, nil, []string{"com.apple.PhotoBooth", "us.zoom.xos"}},
		{"by name", apps, []string{"photo booth"}, []string{"com.apple.PhotoBooth"}},
		{"whole name only", apps, []string{"photo"}, nil},
		{"passthrough id", apps, []string{"org.example.Other"}, []string{"org.example.Other"}},
		{"repeats", apps, []string{"Notes", "com.example.notes", "notes"}, []string{"com.example.Notes"}},
		{"unknown", apps, []string{"nope"}, nil},
		{"no defaults", apps[1:2], nil, nil},
	}
	for _, tt := range tests {
		got, err := selectApps(tt.apps, tt.args)
		if (err != nil) != (tt.want == nil) || !slices.Equal(got, tt.want) {
			t.Errorf("%s: selectApps = %v, %v; want %v", tt.name, got, err, tt.want)
		}
	}
}

func TestPrintStates(t *testing.T) {
	apps := []controlcenter.App{{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth"}}
	states := []controlcenter.State{
		{
			App: "com.apple.PhotoBooth",
			Enabled: map[controlcenter.Effect]bool{
				controlcenter.Portrait: false, controlcenter.StudioLight: true, controlcenter.Reactions: true,
				controlcenter.Background: true,
			},
			Image: "/tmp/a.png",
			Mic:   controlcenter.VoiceIsolation,
			MicOK: true,
		},
		{App: "org.example.Other", Enabled: map[controlcenter.Effect]bool{controlcenter.Background: false}},
	}
	var out bytes.Buffer
	if err := printStates(&out, apps, states); err != nil {
		t.Fatal(err)
	}
	want := "APP                PORTRAIT  STUDIO-LIGHT  REACTIONS  BACKGROUND  IMAGE       MIC-MODE\n" +
		"Photo Booth        off       on            on         on          /tmp/a.png  voice-isolation\n" +
		"org.example.Other  -         -             -          off         -           -\n"
	if out.String() != want {
		t.Errorf("printStates wrote:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestApplyTo(t *testing.T) {
	apps := []controlcenter.App{
		{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth", Known: true},
		{BundleID: "us.zoom.xos", Name: "zoom.us", Known: true},
	}
	current := func(id string) (controlcenter.State, error) {
		return controlcenter.State{App: id, Enabled: map[controlcenter.Effect]bool{controlcenter.Background: id == "us.zoom.xos"}}, nil
	}
	var changed []string
	change := func(ids []string) error { changed = ids; return nil }

	var out bytes.Buffer
	if err := applyTo(&out, apps, []string{"zoom.us", "photo booth"}, change, current); err != nil {
		t.Fatal(err)
	}
	if want := []string{"us.zoom.xos", "com.apple.PhotoBooth"}; !slices.Equal(changed, want) {
		t.Errorf("changed %v, want %v", changed, want)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "zoom.us ") || !strings.HasPrefix(lines[2], "Photo Booth ") {
		t.Errorf("output not in argument order:\n%s", out.String())
	}

	out.Reset()
	failed := errors.New("failed")
	if err := applyTo(&out, apps, nil, func([]string) error { return failed }, current); err != failed {
		t.Errorf("change error: got %v", err)
	}
	if err := applyTo(&out, apps, nil, change, func(string) (controlcenter.State, error) { return controlcenter.State{}, failed }); err != failed {
		t.Errorf("current error: got %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("printed after an error:\n%s", out.String())
	}
}

func TestPrintApps(t *testing.T) {
	apps := []controlcenter.App{
		{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth", Known: true},
		{BundleID: "com.example.Notes", Name: "Notes"},
	}
	var out bytes.Buffer
	if err := printApps(&out, apps); err != nil {
		t.Fatal(err)
	}
	want := "APP          BUNDLE ID             DEFAULT\n" +
		"Photo Booth  com.apple.PhotoBooth  yes\n" +
		"Notes        com.example.Notes     -\n"
	if out.String() != want {
		t.Errorf("printApps wrote:\n%s\nwant:\n%s", out.String(), want)
	}
}
