package controlcenter

import (
	"fmt"
	"slices"
)

// Effect is one switch in Control Center's Video Effects menu.
type Effect int

// The effects, in menu order. Edge Light is missing: see ARCHITECTURE.md.
const (
	Portrait Effect = iota
	StudioLight
	Reactions
	Background
	numEffects
)

// Each name is the menu label in lower case, with hyphens for spaces.
var effectNames = [numEffects]string{
	Portrait:    "portrait",
	StudioLight: "studio-light",
	Reactions:   "reactions",
	Background:  "background",
}

// Effects lists every effect in menu order.
var Effects = [numEffects]Effect{Portrait, StudioLight, Reactions, Background}

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
	Standard       MicMode = 0
	WideSpectrum   MicMode = 1
	VoiceIsolation MicMode = 2
)

// Each name is the menu label in lower case, with hyphens for spaces.
var micNames = [...]string{Standard: "standard", WideSpectrum: "wide-spectrum", VoiceIsolation: "voice-isolation"}

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
