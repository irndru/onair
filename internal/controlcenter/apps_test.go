package controlcenter

import (
	"slices"
	"testing"
)

var testApps = []App{
	{BundleID: "com.apple.PhotoBooth", Name: "Photo Booth", Known: true},
	{BundleID: "com.example.Notes", Name: "Notes"},
	{BundleID: "us.zoom.xos", Name: "zoom.us", Toggled: true},
}

func TestDefaults(t *testing.T) {
	want := []string{"com.apple.PhotoBooth", "us.zoom.xos"}
	if got := Defaults(testApps); !slices.Equal(got, want) {
		t.Errorf("Defaults = %v, want %v", got, want)
	}
	if got := Defaults(nil); got != nil {
		t.Errorf("Defaults(nil) = %v", got)
	}
}

func TestResolveApp(t *testing.T) {
	tests := []struct{ query, want string }{
		{"Photo Booth", "com.apple.PhotoBooth"},
		{"photo booth", "com.apple.PhotoBooth"},
		{"COM.APPLE.PHOTOBOOTH", "com.apple.PhotoBooth"},
		{"zoom.us", "us.zoom.xos"},
		{"org.example.Other", "org.example.Other"},
		{"photo", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got, err := ResolveApp(testApps, tt.query)
		if got != tt.want || (err != nil) != (tt.want == "") {
			t.Errorf("ResolveApp(%q) = %q, %v; want %q", tt.query, got, err, tt.want)
		}
	}
}
