package controlcenter

import "testing"

func TestParseEffect(t *testing.T) {
	for _, e := range Effects {
		if got, ok := ParseEffect(e.String()); !ok || got != e {
			t.Errorf("ParseEffect(%q) = %v, %v", e, got, ok)
		}
	}
	for _, name := range []string{"", "Portrait", "blur", "studio", "edge-light", "mic"} {
		if got, ok := ParseEffect(name); ok || got != 0 {
			t.Errorf("ParseEffect(%q) = %v, %v; want 0, false", name, got, ok)
		}
	}
}

func TestStringOutOfRange(t *testing.T) {
	tests := []struct{ got, want string }{
		{Effect(-1).String(), "Effect(-1)"},
		{numEffects.String(), "Effect(4)"},
		{MicMode(-1).String(), "MicMode(-1)"},
		{MicMode(3).String(), "MicMode(3)"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}

func TestParseMicMode(t *testing.T) {
	for _, m := range []MicMode{Standard, WideSpectrum, VoiceIsolation} {
		if got, ok := ParseMicMode(m.String()); !ok || got != m {
			t.Errorf("ParseMicMode(%q) = %v, %v", m, got, ok)
		}
	}
	for _, name := range []string{"", "Standard", "isolation", "wide"} {
		if got, ok := ParseMicMode(name); ok || got != 0 {
			t.Errorf("ParseMicMode(%q) = %v, %v; want 0, false", name, got, ok)
		}
	}
}
