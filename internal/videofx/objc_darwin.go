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
	effectSupported    func(effect, bundleID objc.ID) bool
	ringLightActive    func(bundleID objc.ID) bool
	setRingLightActive func(on bool, bundleID objc.ID)
	getMicMode         func(bundleID objc.ID) int
	setMicrophoneMode  func(mode int, bundleID objc.ID) bool
	supportedMicModes  func(bundleID objc.ID) objc.ID
)

// Edge has no constant here: Control Center switches it with the ring light
// functions, and the generic ones ignore AVControlCenterVideoEffectRingLight.
var effectSymbols = map[Effect]string{
	Background: "AVControlCenterVideoEffectBackgroundReplacement",
	Portrait:   "AVControlCenterVideoEffectBackgroundBlur",
	Studio:     "AVControlCenterVideoEffectStudioLighting",
	Reactions:  "AVControlCenterVideoEffectReactions",
}

var (
	effectConsts  = map[Effect]objc.ID{}
	effectMissing = map[Effect]string{} // the symbol each unavailable effect lacks
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

// load binds the private AVFoundation functions, once.
var load = sync.OnceValue(func() error {
	lib, err := purego.Dlopen(avFoundation, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	for _, f := range []struct {
		fn   any
		name string
	}{
		{&getBackgroundURL, "AVControlCenterVideoEffectsModuleGetBackgroundReplacementURL"},
		{&setBackgroundURL, "AVControlCenterVideoEffectsModuleSetBackgroundReplacementURL"},
		{&effectEnabled, "AVControlCenterVideoEffectsModuleIsEffectEnabledForBundleID"},
		{&setEffectEnabled, "AVControlCenterVideoEffectsModuleSetEffectEnabledForBundleID"},
		{&backgroundToggled, "AVControlCenterVideoEffectsModuleHasBackgroundReplacementBeenToggledForBundleID"},
	} {
		addr, err := purego.Dlsym(lib, f.name)
		if err != nil {
			return missing(f.name)
		}
		purego.RegisterFunc(f.fn, addr)
	}
	for e, symbol := range effectSymbols {
		addr, err := purego.Dlsym(lib, symbol)
		if err != nil {
			effectMissing[e] = symbol
			continue
		}
		effectConsts[e] = **(**objc.ID)(unsafe.Pointer(&addr))
	}
	if symbol, ok := effectMissing[Background]; ok {
		return missing(symbol)
	}
	if !objc.Send[bool](class("AVCaptureDevice"), selRespondsToSelector, selIsEligible) {
		return missing("+[AVCaptureDevice isEligibleForBackgroundReplacement]")
	}
	// optional binds each function and returns the first symbol it cannot find.
	optional := func(fns ...any) string {
		for i := 0; i < len(fns); i += 2 {
			name := fns[i+1].(string)
			addr, err := purego.Dlsym(lib, name)
			if err != nil {
				return name
			}
			purego.RegisterFunc(fns[i], addr)
		}
		return ""
	}
	optional(&effectSupported, "AVControlCenterVideoEffectsModuleIsEffectSupportedForBundleID")
	if symbol := optional(
		&ringLightActive, "AVControlCenterVideoEffectsModuleGetRingLightActiveForBundleID",
		&setRingLightActive, "AVControlCenterVideoEffectsModuleSetRingLightActiveForBundleID",
	); symbol != "" {
		effectMissing[Edge] = symbol
	}
	micMissing = optional(
		&getMicMode, "AVControlCenterMicrophoneModesModuleGetMicrophoneModeForBundleID",
		&setMicrophoneMode, "AVControlCenterMicrophoneModesModuleSetMicrophoneModeForBundleID",
		&supportedMicModes, "AVControlCenterMicrophoneModesModuleGetSupportedMicrophoneModesForBundleID",
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
	if symbol, ok := effectMissing[e]; ok {
		return missing(symbol)
	}
	return nil
}

func isSupported(e Effect, bundleID string) bool {
	if effectErr(e) != nil {
		return false
	}
	// Edge reports unsupported through the generic function even where it works.
	if e == Edge || effectSupported == nil {
		return true
	}
	return effectSupported(effectConsts[e], nsString(bundleID))
}

func isEnabled(e Effect, bundleID string) bool {
	if e == Edge {
		return ringLightActive(nsString(bundleID))
	}
	return effectEnabled(effectConsts[e], nsString(bundleID))
}

func setEnabled(e Effect, on bool, bundleID string) {
	if e == Edge {
		setRingLightActive(on, nsString(bundleID))
		return
	}
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
