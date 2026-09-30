package videofx

import (
	"fmt"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const avFoundation = "/System/Library/Frameworks/AVFoundation.framework/AVFoundation"

var (
	getBackgroundURL  func(bundleID objc.ID) objc.ID
	setBackgroundURL  func(url, bundleID objc.ID)
	effectEnabled     func(effect, bundleID objc.ID) bool
	setEffectEnabled  func(effect objc.ID, on bool, bundleID objc.ID)
	backgroundToggled func(bundleID objc.ID) bool

	// Optional: nil when this macOS version lacks the symbol.
	effectSupported   func(effect, bundleID objc.ID) bool
	getMicMode        func(bundleID objc.ID) int
	setMicrophoneMode func(mode int, bundleID objc.ID) bool
	supportedMicModes func(bundleID objc.ID) objc.ID
)

// The menu's Reactions switch is the Gestures effect. The Reactions effect
// is a different setting that the menu does not show.
var effectSymbols = [numEffects]string{
	Background: "AVControlCenterVideoEffectBackgroundReplacement",
	Portrait:   "AVControlCenterVideoEffectBackgroundBlur",
	Studio:     "AVControlCenterVideoEffectStudioLighting",
	Reactions:  "AVControlCenterVideoEffectGestures",
}

var (
	effectConsts  [numEffects]objc.ID
	effectMissing [numEffects]string // the symbol each unavailable effect lacks
	micMissing    string
)

var (
	selStringWithUTF8String = objc.RegisterName("stringWithUTF8String:")
	selUTF8String           = objc.RegisterName("UTF8String")
	selFileURLWithPath      = objc.RegisterName("fileURLWithPath:")
	selPath                 = objc.RegisterName("path")
	selDictionaryWithFile   = objc.RegisterName("dictionaryWithContentsOfFile:")
	selObjectForKey         = objc.RegisterName("objectForKey:")
	selIsKindOfClass        = objc.RegisterName("isKindOfClass:")
	selRespondsToSelector   = objc.RegisterName("respondsToSelector:")
	selIsEligible           = objc.RegisterName("isEligibleForBackgroundReplacement")
	selCount                = objc.RegisterName("count")
	selObjectAtIndex        = objc.RegisterName("objectAtIndex:")
	selIntegerValue         = objc.RegisterName("integerValue")
)

func missing(symbol string) error {
	return fmt.Errorf("this macOS version is not supported: AVFoundation has no %s", symbol)
}

// binding pairs a pointer to a func variable with the C symbol it calls.
type binding struct {
	fn   any
	name string
}

// bind registers each binding and returns the first symbol it cannot find.
func bind(lib uintptr, bs ...binding) string {
	for _, b := range bs {
		addr, err := purego.Dlsym(lib, b.name)
		if err != nil {
			return b.name
		}
		purego.RegisterFunc(b.fn, addr)
	}
	return ""
}

// load binds the private AVFoundation functions, once.
var load = sync.OnceValue(func() error {
	lib, err := purego.Dlopen(avFoundation, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("open AVFoundation: %w", err)
	}
	if symbol := bind(lib,
		binding{&getBackgroundURL, "AVControlCenterVideoEffectsModuleGetBackgroundReplacementURL"},
		binding{&setBackgroundURL, "AVControlCenterVideoEffectsModuleSetBackgroundReplacementURL"},
		binding{&effectEnabled, "AVControlCenterVideoEffectsModuleIsEffectEnabledForBundleID"},
		binding{&setEffectEnabled, "AVControlCenterVideoEffectsModuleSetEffectEnabledForBundleID"},
		binding{&backgroundToggled, "AVControlCenterVideoEffectsModuleHasBackgroundReplacementBeenToggledForBundleID"},
	); symbol != "" {
		return missing(symbol)
	}
	for e, symbol := range effectSymbols {
		addr, err := purego.Dlsym(lib, symbol)
		if err != nil {
			effectMissing[e] = symbol
			continue
		}
		// addr is the C address of an NSString *const. Read the pointer stored there.
		effectConsts[e] = **(**objc.ID)(unsafe.Pointer(&addr))
	}
	if symbol := effectMissing[Background]; symbol != "" {
		return missing(symbol)
	}
	if !objc.Send[bool](class("AVCaptureDevice"), selRespondsToSelector, selIsEligible) {
		return missing("+[AVCaptureDevice isEligibleForBackgroundReplacement]")
	}
	bind(lib, binding{&effectSupported, "AVControlCenterVideoEffectsModuleIsEffectSupportedForBundleID"})
	micMissing = bind(lib,
		binding{&getMicMode, "AVControlCenterMicrophoneModesModuleGetMicrophoneModeForBundleID"},
		binding{&setMicrophoneMode, "AVControlCenterMicrophoneModesModuleSetMicrophoneModeForBundleID"},
		binding{&supportedMicModes, "AVControlCenterMicrophoneModesModuleGetSupportedMicrophoneModesForBundleID"},
	)
	return nil
})

func class(name string) objc.ID { return objc.ID(objc.GetClass(name)) }

func nsString(s string) objc.ID {
	return class("NSString").Send(selStringWithUTF8String, s)
}

func goString(s objc.ID) string {
	if s == 0 {
		return ""
	}
	return objc.Send[string](s, selUTF8String)
}

func isEligible() bool {
	return objc.Send[bool](class("AVCaptureDevice"), selIsEligible)
}

func getURL(bundleID string) string {
	url := getBackgroundURL(nsString(bundleID))
	if url == 0 {
		return ""
	}
	return goString(url.Send(selPath))
}

func setURL(path, bundleID string) {
	setBackgroundURL(class("NSURL").Send(selFileURLWithPath, nsString(path)), nsString(bundleID))
}

// effectErr reports why this macOS version cannot switch e, if it cannot.
func effectErr(e Effect) error {
	if symbol := effectMissing[e]; symbol != "" {
		return missing(symbol)
	}
	return nil
}

func isSupported(e Effect, bundleID string) bool {
	if effectErr(e) != nil {
		return false
	}
	if effectSupported == nil {
		return true
	}
	return effectSupported(effectConsts[e], nsString(bundleID))
}

func isEnabled(e Effect, bundleID string) bool {
	return effectEnabled(effectConsts[e], nsString(bundleID))
}

func setEnabled(e Effect, on bool, bundleID string) {
	setEffectEnabled(effectConsts[e], on, nsString(bundleID))
}

func micErr() error {
	if micMissing != "" {
		return missing(micMissing)
	}
	return nil
}

func micMode(bundleID string) MicMode {
	return MicMode(getMicMode(nsString(bundleID)))
}

// micModes lists the modes the app supports. Setting any other mode raises
// an Objective-C exception.
func micModes(bundleID string) []MicMode {
	list := supportedMicModes(nsString(bundleID))
	if list == 0 {
		return nil
	}
	var modes []MicMode
	for i := range objc.Send[int](list, selCount) {
		modes = append(modes, MicMode(objc.Send[int](list.Send(selObjectAtIndex, i), selIntegerValue)))
	}
	return modes
}

func setMicMode(mode MicMode, bundleID string) {
	setMicrophoneMode(int(mode), nsString(bundleID))
}

func toggled(bundleID string) bool {
	return backgroundToggled(nsString(bundleID))
}

func bundleInfo(appPath string) (bundle, bool) {
	dict := class("NSDictionary").Send(selDictionaryWithFile, nsString(filepath.Join(appPath, "Contents", "Info.plist")))
	if dict == 0 {
		return bundle{}, false
	}
	str := func(key string) string {
		v := dict.Send(selObjectForKey, nsString(key))
		if v == 0 || !objc.Send[bool](v, selIsKindOfClass, class("NSString")) {
			return ""
		}
		return goString(v)
	}
	b := bundle{
		id:          str("CFBundleIdentifier"),
		displayName: str("CFBundleDisplayName"),
		name:        str("CFBundleName"),
		camera:      dict.Send(selObjectForKey, nsString("NSCameraUsageDescription")) != 0,
	}
	return b, b.id != ""
}
