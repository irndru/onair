package controlcenter

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestChangeRejectsBadIDs(t *testing.T) {
	f := useFake(t, &fakeBridge{})
	for _, apps := range [][]string{{""}, {"a.one", "a b"}, {"tab\tid"}} {
		if err := SetEnabled(Portrait, true, apps...); err == nil {
			t.Errorf("SetEnabled(%q) succeeded", apps)
		}
	}
	if f.calls != nil {
		t.Errorf("calls = %q, want none", f.calls)
	}
}

// A check that fails for any app must stop the change for every app.
func TestChangeIsAllOrNothing(t *testing.T) {
	tests := []struct {
		name string
		f    *fakeBridge
		set  func(apps ...string) error
		want string // part of the error
	}{
		{
			"unsupported effect",
			&fakeBridge{apps: map[string]*fakeApp{"a.two": {unsupported: []Effect{Portrait}}}},
			func(apps ...string) error { return SetEnabled(Portrait, true, apps...) },
			"a.two does not support the portrait effect",
		},
		{
			"missing effect symbol",
			&fakeBridge{effectErrs: map[Effect]error{StudioLight: errors.New("no studio-light symbol")}},
			func(apps ...string) error { return SetEnabled(StudioLight, true, apps...) },
			"no studio-light symbol",
		},
		{
			"not eligible",
			&fakeBridge{notEligible: true},
			func(apps ...string) error { return SetImage("/p.png", apps...) },
			ErrNotEligible.Error(),
		},
		{
			"missing mic symbols",
			&fakeBridge{micMissing: errors.New("no mic symbol")},
			func(apps ...string) error { return SetMic(VoiceIsolation, apps...) },
			"no mic symbol",
		},
		{
			// Setting an unsupported mode raises an Objective-C exception.
			"unsupported mic mode",
			&fakeBridge{apps: map[string]*fakeApp{
				"a.one":   {micModes: []MicMode{Standard, VoiceIsolation}},
				"a.two":   {micModes: []MicMode{Standard}},
				"a.three": {micModes: []MicMode{Standard, VoiceIsolation}},
			}},
			func(apps ...string) error { return SetMic(VoiceIsolation, apps...) },
			"a.two does not support mic mode voice-isolation",
		},
	}
	for _, tt := range tests {
		f := useFake(t, tt.f)
		err := tt.set("a.one", "a.two", "a.three")
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want %q", tt.name, err, tt.want)
		}
		if f.calls != nil {
			t.Errorf("%s: calls = %q, want none", tt.name, f.calls)
		}
	}
}

func TestNotEligibleOnlyBlocksBackground(t *testing.T) {
	f := useFake(t, &fakeBridge{notEligible: true})
	if err := SetEnabled(Background, true, "a.one"); !errors.Is(err, ErrNotEligible) {
		t.Errorf("SetEnabled(Background) = %v, want ErrNotEligible", err)
	}
	if err := SetEnabled(Portrait, true, "a.one"); err != nil {
		t.Errorf("SetEnabled(Portrait) = %v", err)
	}
	if want := []string{"setEnabled portrait true a.one"}; !slices.Equal(f.calls, want) {
		t.Errorf("calls = %q, want %q", f.calls, want)
	}
}

func TestSetters(t *testing.T) {
	both := &fakeApp{micModes: []MicMode{Standard, VoiceIsolation}}
	tests := []struct {
		name string
		set  func() error
		want []string
	}{
		{
			"SetImage",
			func() error { return SetImage("/p.png", "a.one", "a.two") },
			[]string{
				"setURL /p.png a.one", "setEnabled background true a.one",
				"setURL /p.png a.two", "setEnabled background true a.two",
			},
		},
		{
			"SetEnabled",
			func() error { return SetEnabled(Reactions, false, "a.two", "a.one") },
			[]string{"setEnabled reactions false a.two", "setEnabled reactions false a.one"},
		},
		{
			"SetMic",
			func() error { return SetMic(VoiceIsolation, "a.one", "a.two") },
			[]string{"setMic voice-isolation a.one", "setMic voice-isolation a.two"},
		},
	}
	for _, tt := range tests {
		f := useFake(t, &fakeBridge{apps: map[string]*fakeApp{"a.one": both, "a.two": both}})
		if err := tt.set(); err != nil {
			t.Errorf("%s: %v", tt.name, err)
		}
		if !slices.Equal(f.calls, tt.want) {
			t.Errorf("%s: calls = %q, want %q", tt.name, f.calls, tt.want)
		}
	}
}

func TestCurrent(t *testing.T) {
	f := &fakeBridge{apps: map[string]*fakeApp{
		"a.one": {
			unsupported: []Effect{StudioLight},
			enabled:     map[Effect]bool{Background: true, StudioLight: true},
			image:       "/p.png",
			mic:         VoiceIsolation,
			micModes:    []MicMode{Standard, VoiceIsolation},
		},
		"a.stale":  {mic: WideSpectrum, micModes: []MicMode{Standard}},
		"a.future": {mic: MicMode(3), micModes: []MicMode{Standard, MicMode(3)}},
	}}
	useFake(t, f)
	tests := []struct {
		app  string
		want State
	}{
		{"a.one", State{
			App:     "a.one",
			Enabled: map[Effect]bool{Background: true, Portrait: false, Reactions: false},
			Image:   "/p.png",
			Mic:     VoiceIsolation,
			MicOK:   true,
		}},
		// The recorded mode is not one the app supports, so it is not shown.
		{"a.stale", State{App: "a.stale", Mic: WideSpectrum}},
		{"a.future", State{App: "a.future", Mic: MicMode(3), MicOK: true}},
	}
	for _, tt := range tests {
		got, err := Current(tt.app)
		if err != nil {
			t.Fatal(err)
		}
		if tt.want.Enabled == nil {
			tt.want.Enabled = map[Effect]bool{Background: false, Portrait: false, StudioLight: false, Reactions: false}
		}
		if !maps.Equal(got.Enabled, tt.want.Enabled) || got.Image != tt.want.Image ||
			got.Mic != tt.want.Mic || got.MicOK != tt.want.MicOK || got.App != tt.want.App {
			t.Errorf("Current(%q) = %+v, want %+v", tt.app, got, tt.want)
		}
	}

	f.micMissing = errors.New("no mic symbol")
	if got, _ := Current("a.one"); got.MicOK {
		t.Errorf("Current with no mic symbols has MicOK")
	}
}

func TestOpenBridgeError(t *testing.T) {
	old := openBridge
	openBridge = func() (bridge, error) { return nil, errors.New("no bridge") }
	t.Cleanup(func() { openBridge = old })
	if _, err := Current("a.one"); err == nil {
		t.Error("Current succeeded without a bridge")
	}
	if err := SetMic(Standard, "a.one"); err == nil {
		t.Error("SetMic succeeded without a bridge")
	}
	if _, err := appsIn(nil); err == nil {
		t.Error("appsIn succeeded without a bridge")
	}
}

func TestSetMicRefused(t *testing.T) {
	f := useFake(t, &fakeBridge{
		refuseMic: true,
		apps:      map[string]*fakeApp{"a.one": {micModes: []MicMode{Standard}}, "a.two": {micModes: []MicMode{Standard}}},
	})
	err := SetMic(Standard, "a.one", "a.two")
	if err == nil || !strings.Contains(err.Error(), "refused mic mode standard for a.one") {
		t.Errorf("err = %v", err)
	}
	// It stops at the first refusal.
	if want := []string{"setMic standard a.one"}; !slices.Equal(f.calls, want) {
		t.Errorf("calls = %q, want %q", f.calls, want)
	}
}
