// Package videofx reads and changes the macOS video effects and mic mode
// that Control Center keeps for each app.
package videofx

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type State struct {
	App     string
	Enabled map[Effect]bool // holds only the effects the app supports
	Image   string
	Mic     MicMode
	MicOK   bool // false when the app has no mic mode
}

var ErrNotEligible = errors.New("this Mac does not support camera backgrounds")

func Current(app string) (State, error) {
	if err := load(); err != nil {
		return State{}, err
	}
	s := State{App: app, Enabled: map[Effect]bool{}, Image: getURL(app)}
	for _, e := range Effects {
		if isSupported(e, app) {
			s.Enabled[e] = isEnabled(e, app)
		}
	}
	if micErr() == nil {
		s.Mic = micMode(app)
		s.MicOK = slices.Contains(micModes(app), s.Mic)
	}
	return s, nil
}

// SetImage sets the background image and turns the effect on.
func SetImage(path string, apps ...string) error {
	return change(apps, checkEffect(Background), func(app string) {
		setURL(path, app)
		setEnabled(Background, true, app)
	})
}

func SetEnabled(e Effect, on bool, apps ...string) error {
	return change(apps, checkEffect(e), func(app string) { setEnabled(e, on, app) })
}

func SetMic(mode MicMode, apps ...string) error {
	check := func(app string) error {
		if err := micErr(); err != nil {
			return err
		}
		if !slices.Contains(micModes(app), mode) {
			return fmt.Errorf("%s does not support mic mode %s", app, mode)
		}
		return nil
	}
	return change(apps, check, func(app string) { setMicMode(mode, app) })
}

func checkEffect(e Effect) func(app string) error {
	return func(app string) error {
		if e == Background && !isEligible() {
			return ErrNotEligible
		}
		if err := effectErr(e); err != nil {
			return err
		}
		if !isSupported(e, app) {
			return fmt.Errorf("%s does not support the %s effect", app, e)
		}
		return nil
	}
}

// change applies to every app, or to none when check fails for any of them.
func change(apps []string, check func(app string) error, apply func(app string)) error {
	for _, app := range apps {
		if app == "" || strings.ContainsAny(app, " \t\n") {
			return fmt.Errorf("invalid bundle identifier %q", app)
		}
	}
	if err := load(); err != nil {
		return err
	}
	for _, app := range apps {
		if err := check(app); err != nil {
			return err
		}
	}
	for _, app := range apps {
		apply(app)
	}
	return nil
}
