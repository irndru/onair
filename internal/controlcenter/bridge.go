package controlcenter

import "sync"

// bridge is the system side of controlcenter: the private AVFoundation
// functions and Info.plist reads. Tests swap in a fake so they never reach the
// real setters.
type bridge interface {
	eligible() bool
	effectErr(e Effect) error // why this macOS cannot switch e, if it cannot
	supported(e Effect, app string) bool
	enabled(e Effect, app string) bool
	setEnabled(e Effect, on bool, app string)
	url(app string) string
	setURL(path, app string)
	micErr() error // why this macOS has no mic mode, if it has none
	mic(app string) MicMode
	micModes(app string) []MicMode
	setMic(m MicMode, app string) bool // false when the daemon refuses
	toggled(app string) bool
	bundleInfo(appPath string) (bundle, bool)
}

// openBridge loads the bridge, once.
var openBridge = sync.OnceValues(newBridge)
