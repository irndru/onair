//go:build !darwin

package videofx

import "errors"

func load() error { return errors.New("videobg only runs on macOS") }

func isEligible() bool                 { return false }
func getURL(string) string             { return "" }
func setURL(string, string)            {}
func isEnabled(string) bool            { return false }
func setEnabled(bool, string)          {}
func toggled(string) bool              { return false }
func bundleInfo(string) (bundle, bool) { return bundle{}, false }
