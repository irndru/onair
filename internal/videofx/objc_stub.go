//go:build !darwin

package videofx

import "errors"

func load() error { return errors.New("videobg only runs on macOS") }

func isEligible() bool                 { return false }
func getURL(string) string             { return "" }
func setURL(string, string)            {}
func effectErr(Effect) error           { return nil }
func isSupported(Effect, string) bool  { return false }
func isEnabled(Effect, string) bool    { return false }
func setEnabled(Effect, bool, string)  {}
func micErr() error                    { return nil }
func micMode(string) MicMode           { return Standard }
func micModes(string) []MicMode        { return nil }
func setMicMode(MicMode, string)       {}
func toggled(string) bool              { return false }
func bundleInfo(string) (bundle, bool) { return bundle{}, false }
