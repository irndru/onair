package videofx

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const gradientDir = "/System/Library/PrivateFrameworks/Portrait.framework/Resources/backgrounds/gradient"

var gradients = map[string]string{
	"blue":      "1_blue.png",
	"purple":    "2_purple.png",
	"pink":      "3_pink.png",
	"red":       "4_red.png",
	"orange":    "5_orange.png",
	"yellow":    "6_yellow.png",
	"green":     "7_green.png",
	"off-white": "8_off_white.png",
	"dark":      "9_dark.png",
}

// Builtin returns the names of Apple's built-in backgrounds.
func Builtin() []string { return slices.Sorted(maps.Keys(gradients)) }

func BuiltinPath(name string) (string, bool) {
	file, ok := gradients[strings.ToLower(name)]
	return filepath.Join(gradientDir, file), ok
}

// ResolveImage turns a built-in name or a file path into an absolute path.
func ResolveImage(nameOrPath string) (string, error) {
	path, ok := BuiltinPath(nameOrPath)
	if !ok {
		var err error
		if path, err = filepath.Abs(nameOrPath); err != nil {
			return "", err
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a file", path)
	}
	return path, nil
}
