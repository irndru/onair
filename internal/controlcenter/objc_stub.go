//go:build !darwin

package controlcenter

import "errors"

func newBridge() (bridge, error) { return nil, errors.New("onair only runs on macOS") }
