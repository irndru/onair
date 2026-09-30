package videofx

import "slices"

// Effect is one switch in Control Center's Video Effects menu.
type Effect int

const (
	Background Effect = iota
	Portrait
	Studio
	Reactions
)

var effectNames = []string{"background", "portrait", "studio", "reactions"}

// Effects lists every effect in display order.
var Effects = []Effect{Background, Portrait, Studio, Reactions}

func (e Effect) String() string { return effectNames[e] }

func ParseEffect(name string) (Effect, bool) {
	i := slices.Index(effectNames, name)
	return Effect(i), i >= 0
}

// MicMode values match AVCaptureMicrophoneMode.
type MicMode int

const (
	Standard  MicMode = 0
	Wide      MicMode = 1
	Isolation MicMode = 2
)

var micNames = []string{"standard", "wide", "isolation"}

func (m MicMode) String() string { return micNames[m] }

func ParseMicMode(name string) (MicMode, bool) {
	i := slices.Index(micNames, name)
	return MicMode(i), i >= 0
}
