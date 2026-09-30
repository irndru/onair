//go:build !darwin

package videofx

import "errors"

func newBridge() (bridge, error) { return nil, errors.New("videobg only runs on macOS") }
