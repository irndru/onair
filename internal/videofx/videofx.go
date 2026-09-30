// Package videofx reads and changes the macOS camera background that
// Control Center keeps for each app.
package videofx

import (
	"errors"
	"fmt"
	"strings"
)

type State struct {
	App     string
	Enabled bool
	Image   string
}

var ErrNotEligible = errors.New("this Mac does not support camera backgrounds")

func Current(app string) (State, error) {
	if err := load(); err != nil {
		return State{}, err
	}
	return State{App: app, Enabled: isEnabled(app), Image: getURL(app)}, nil
}

// SetImage sets the background image and turns the effect on.
func SetImage(path string, apps ...string) error {
	return change(apps, func(app string) {
		setURL(path, app)
		setEnabled(true, app)
	})
}

func SetEnabled(on bool, apps ...string) error {
	return change(apps, func(app string) { setEnabled(on, app) })
}

func change(apps []string, apply func(app string)) error {
	for _, app := range apps {
		if app == "" || strings.ContainsAny(app, " \t\n") {
			return fmt.Errorf("invalid bundle identifier %q", app)
		}
	}
	if err := load(); err != nil {
		return err
	}
	if !isEligible() {
		return ErrNotEligible
	}
	for _, app := range apps {
		apply(app)
	}
	return nil
}
