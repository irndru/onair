package main

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/AndrewMcCraeCA/videobg/internal/videofx"
)

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"bogus"}, {"set"}} {
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
	if !strings.HasPrefix(out.String(), "videobg ") {
		t.Errorf("version wrote %q", out.String())
	}
}

func TestBackgrounds(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"backgrounds"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, name := range videofx.Builtin() {
		if !strings.Contains(out.String(), name+" ") {
			t.Errorf("backgrounds does not list %q", name)
		}
	}
}

func TestSetMissingImage(t *testing.T) {
	err := run([]string{"set", filepath.Join(t.TempDir(), "missing.png")}, io.Discard)
	var usageErr usageError
	if err == nil || errors.As(err, &usageErr) {
		t.Errorf("err = %v, want a non-usage error", err)
	}
}

func TestSelectApps(t *testing.T) {
	apps := []videofx.App{
		{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth", Known: true},
		{BundleID: "com.example.Notes", Name: "Notes"},
		{BundleID: "us.zoom.xos", Name: "zoom.us", Toggled: true},
	}
	tests := []struct {
		name string
		apps []videofx.App
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
	apps := []videofx.App{{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth"}}
	states := []videofx.State{
		{App: "com.apple.PhotoBooth", Enabled: true, Image: "/tmp/a.png"},
		{App: "org.example.Other"},
	}
	var out bytes.Buffer
	if err := printStates(&out, apps, states); err != nil {
		t.Fatal(err)
	}
	want := "Photo Booth        on   /tmp/a.png\n" +
		"org.example.Other  off  -\n"
	if out.String() != want {
		t.Errorf("printStates wrote:\n%s\nwant:\n%s", out.String(), want)
	}
}
