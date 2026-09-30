package videofx

import "testing"

func TestParseEffect(t *testing.T) {
	for _, e := range Effects {
		if got, ok := ParseEffect(e.String()); !ok || got != e {
			t.Errorf("ParseEffect(%q) = %v, %v", e, got, ok)
		}
	}
	for _, name := range []string{"", "Portrait", "blur", "mic"} {
		if got, ok := ParseEffect(name); ok {
			t.Errorf("ParseEffect(%q) = %v, want no match", name, got)
		}
	}
}

func TestParseMicMode(t *testing.T) {
	for _, m := range []MicMode{Standard, Wide, Isolation} {
		if got, ok := ParseMicMode(m.String()); !ok || got != m {
			t.Errorf("ParseMicMode(%q) = %v, %v", m, got, ok)
		}
	}
	for _, name := range []string{"", "Standard", "voice"} {
		if got, ok := ParseMicMode(name); ok {
			t.Errorf("ParseMicMode(%q) = %v, want no match", name, got)
		}
	}
}
