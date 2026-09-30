package videofx

import (
	"fmt"
	"slices"
)

// Effect is one switch in Control Center's Video Effects menu.
type Effect int

// The effects, in display order.
const (
	Background Effect = iota
	Portrait
	Studio
	Reactions
	numEffects
)

var effectNames = [numEffects]string{
	Background: "background",
	Portrait:   "portrait",
	Studio:     "studio",
	Reactions:  "reactions",
}

// Effects lists every effect in display order.
var Effects = [numEffects]Effect{Background, Portrait, Studio, Reactions}

// String returns the effect's command name.
func (e Effect) String() string {
	if e < 0 || e >= numEffects {
		return fmt.Sprintf("Effect(%d)", int(e))
	}
	return effectNames[e]
}

// ParseEffect returns the effect with the given command name.
func ParseEffect(name string) (Effect, bool) {
	i := slices.Index(effectNames[:], name)
	if i < 0 {
		return 0, false
	}
	return Effect(i), true
}

// MicMode is a Mic Mode setting. The values match AVCaptureMicrophoneMode.
type MicMode int

// The mic modes.
const (
	Standard  MicMode = 0
	Wide      MicMode = 1
	Isolation MicMode = 2
)

var micNames = [...]string{Standard: "standard", Wide: "wide", Isolation: "isolation"}

// String returns the mode's command name.
func (m MicMode) String() string {
	if m < 0 || int(m) >= len(micNames) {
		return fmt.Sprintf("MicMode(%d)", int(m))
	}
	return micNames[m]
}

// ParseMicMode returns the mode with the given command name.
func ParseMicMode(name string) (MicMode, bool) {
	i := slices.Index(micNames[:], name)
	if i < 0 {
		return 0, false
	}
	return MicMode(i), true
}
