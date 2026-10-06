package controlcenter

import (
	"fmt"
	"slices"
	"testing"
)

// fakeApp is what fakeBridge has recorded for one app.
type fakeApp struct {
	unsupported []Effect
	enabled     map[Effect]bool
	image       string
	mic         MicMode
	micModes    []MicMode
	toggled     bool
}

// fakeBridge stands in for AVFoundation. It records every set call.
type fakeBridge struct {
	notEligible bool
	effectErrs  map[Effect]error
	micMissing  error
	apps        map[string]*fakeApp
	bundles     map[string]bundle // by app path
	refuseMic   bool
	calls       []string
}

// useFake makes openBridge return f for the rest of the test.
func useFake(t *testing.T, f *fakeBridge) *fakeBridge {
	t.Helper()
	old := openBridge
	openBridge = func() (bridge, error) { return f, nil }
	t.Cleanup(func() { openBridge = old })
	return f
}

func (f *fakeBridge) app(app string) *fakeApp {
	if a, ok := f.apps[app]; ok {
		return a
	}
	return &fakeApp{}
}

func (f *fakeBridge) record(format string, args ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, args...))
}

func (f *fakeBridge) eligible() bool           { return !f.notEligible }
func (f *fakeBridge) effectErr(e Effect) error { return f.effectErrs[e] }
func (f *fakeBridge) enabled(e Effect, app string) bool {
	return f.app(app).enabled[e]
}
func (f *fakeBridge) supported(e Effect, app string) bool {
	return f.effectErrs[e] == nil && !slices.Contains(f.app(app).unsupported, e)
}
func (f *fakeBridge) setEnabled(e Effect, on bool, app string) {
	f.record("setEnabled %s %v %s", e, on, app)
}
func (f *fakeBridge) url(app string) string         { return f.app(app).image }
func (f *fakeBridge) setURL(path, app string)       { f.record("setURL %s %s", path, app) }
func (f *fakeBridge) micErr() error                 { return f.micMissing }
func (f *fakeBridge) mic(app string) MicMode        { return f.app(app).mic }
func (f *fakeBridge) micModes(app string) []MicMode { return f.app(app).micModes }
func (f *fakeBridge) setMic(m MicMode, app string) bool {
	f.record("setMic %s %s", m, app)
	return !f.refuseMic
}
func (f *fakeBridge) toggled(app string) bool { return f.app(app).toggled }
func (f *fakeBridge) bundleInfo(path string) (bundle, bool) {
	b, ok := f.bundles[path]
	return b, ok
}
