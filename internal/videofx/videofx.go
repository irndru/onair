// Package videofx reads and changes the macOS video effects and mic mode
// that Control Center keeps for each app.
package videofx

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// State is what Control Center has recorded for one app.
type State struct {
	App     string
	Enabled map[Effect]bool // holds only the effects the app supports
	Image   string
	Mic     MicMode
	MicOK   bool // false when the app has no mic mode
}

// ErrNotEligible means the Mac cannot replace the camera background at all.
var ErrNotEligible = errors.New("this Mac does not support camera backgrounds")

// Current reads the state of one app.
func Current(app string) (State, error) {
	b, err := openBridge()
	if err != nil {
		return State{}, err
	}
	s := State{App: app, Enabled: map[Effect]bool{}, Image: b.url(app)}
	for _, e := range Effects {
		if b.supported(e, app) {
			s.Enabled[e] = b.enabled(e, app)
		}
	}
	var modes []MicMode
	if b.micErr() == nil {
		s.Mic = b.mic(app)
		modes = b.micModes(app)
		s.MicOK = slices.Contains(modes, s.Mic)
	}
	slog.Debug("read", "app", app, "enabled", s.Enabled, "image", s.Image, "mic", int(s.Mic), "micModes", modes)
	return s, nil
}

// SetImage sets the background image and turns the effect on.
func SetImage(path string, apps ...string) error {
	return change(apps, checkEffect(Background), func(b bridge, app string) error {
		slog.Debug("set image", "app", app, "path", path)
		b.setURL(path, app)
		b.setEnabled(Background, true, app)
		return nil
	})
}

// SetEnabled turns an effect on or off.
func SetEnabled(e Effect, on bool, apps ...string) error {
	return change(apps, checkEffect(e), func(b bridge, app string) error {
		slog.Debug("set effect", "app", app, "effect", e, "on", on)
		b.setEnabled(e, on, app)
		return nil
	})
}

// SetMic sets the mic mode.
func SetMic(mode MicMode, apps ...string) error {
	check := func(b bridge, app string) error {
		if err := b.micErr(); err != nil {
			return err
		}
		if !slices.Contains(b.micModes(app), mode) {
			return fmt.Errorf("%s does not support mic mode %s", app, mode)
		}
		return nil
	}
	return change(apps, check, func(b bridge, app string) error {
		ok := b.setMic(mode, app)
		slog.Debug("set mic", "app", app, "mode", mode, "ok", ok)
		if !ok {
			return fmt.Errorf("macOS refused mic mode %s for %s", mode, app)
		}
		return nil
	})
}

func checkEffect(e Effect) func(b bridge, app string) error {
	return func(b bridge, app string) error {
		if e == Background && !b.eligible() {
			return ErrNotEligible
		}
		if err := b.effectErr(e); err != nil {
			return err
		}
		if !b.supported(e, app) {
			return fmt.Errorf("%s does not support the %s effect", app, e)
		}
		return nil
	}
}

// change applies to every app, or to none when check fails for any of them.
// It stops at the first apply that fails. Only the mic setter reports
// failure, so callers read the state back to confirm the rest.
func change(apps []string, check func(b bridge, app string) error, apply func(b bridge, app string) error) error {
	for _, app := range apps {
		if app == "" || strings.ContainsAny(app, " \t\n") {
			return fmt.Errorf("invalid bundle identifier %q", app)
		}
	}
	b, err := openBridge()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if err := check(b, app); err != nil {
			slog.Debug("check failed, changing nothing", "app", app, "err", err)
			return err
		}
	}
	for _, app := range apps {
		if err := apply(b, app); err != nil {
			return err
		}
	}
	return nil
}
