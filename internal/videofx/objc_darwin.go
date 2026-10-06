package videofx

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const avFoundation = "/System/Library/Frameworks/AVFoundation.framework/AVFoundation"

// The menu's Reactions switch is the Gestures effect. The Reactions effect
// is a different setting that the menu does not show.
var effectSymbols = [numEffects]string{
	Portrait:    "AVControlCenterVideoEffectBackgroundBlur",
	StudioLight: "AVControlCenterVideoEffectStudioLighting",
	Reactions:   "AVControlCenterVideoEffectGestures",
	Background:  "AVControlCenterVideoEffectBackgroundReplacement",
}

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

// avf is the bridge to AVFoundation and Foundation.
type avf struct {
	getBackgroundURL  func(bundleID objc.ID) objc.ID
	setBackgroundURL  func(url, bundleID objc.ID)
	isEffectEnabled   func(effect, bundleID objc.ID) bool
	setEffectEnabled  func(effect objc.ID, on bool, bundleID objc.ID)
	backgroundToggled func(bundleID objc.ID) bool

	// Optional: nil when this macOS version lacks the symbol.
	isEffectSupported func(effect, bundleID objc.ID) bool
	getMicMode        func(bundleID objc.ID) int
	setMicMode        func(mode int, bundleID objc.ID) bool
	supportedMicModes func(bundleID objc.ID) objc.ID

	effects       [numEffects]objc.ID
	effectMissing [numEffects]string // the symbol each unavailable effect lacks
	micMissing    string
}

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
			slog.Debug("symbol missing", "name", b.name)
			return b.name
		}
		purego.RegisterFunc(b.fn, addr)
		slog.Debug("symbol bound", "name", b.name)
	}
	return ""
}

// newBridge binds the private AVFoundation functions.
func newBridge() (bridge, error) {
	version, _ := syscall.Sysctl("kern.osproductversion")
	slog.Debug("opening AVFoundation", "macos", version, "path", avFoundation)
	lib, err := purego.Dlopen(avFoundation, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, fmt.Errorf("open AVFoundation: %w", err)
	}
	a := &avf{}
	if symbol := bind(lib,
		binding{&a.getBackgroundURL, "AVControlCenterVideoEffectsModuleGetBackgroundReplacementURL"},
		binding{&a.setBackgroundURL, "AVControlCenterVideoEffectsModuleSetBackgroundReplacementURL"},
		binding{&a.isEffectEnabled, "AVControlCenterVideoEffectsModuleIsEffectEnabledForBundleID"},
		binding{&a.setEffectEnabled, "AVControlCenterVideoEffectsModuleSetEffectEnabledForBundleID"},
		binding{&a.backgroundToggled, "AVControlCenterVideoEffectsModuleHasBackgroundReplacementBeenToggledForBundleID"},
	); symbol != "" {
		return nil, missing(symbol)
	}
	for e, symbol := range effectSymbols {
		addr, err := purego.Dlsym(lib, symbol)
		if err != nil {
			slog.Debug("effect missing", "effect", Effect(e), "symbol", symbol)
			a.effectMissing[e] = symbol
			continue
		}
		// addr is the C address of an NSString *const. Read the pointer stored there.
		a.effects[e] = **(**objc.ID)(unsafe.Pointer(&addr))
		slog.Debug("effect bound", "effect", Effect(e), "value", goString(a.effects[e]))
	}
	if symbol := a.effectMissing[Background]; symbol != "" {
		return nil, missing(symbol)
	}
	if !objc.Send[bool](class("AVCaptureDevice"), selRespondsToSelector, selIsEligible) {
		return nil, missing("+[AVCaptureDevice isEligibleForBackgroundReplacement]")
	}
	bind(lib, binding{&a.isEffectSupported, "AVControlCenterVideoEffectsModuleIsEffectSupportedForBundleID"})
	a.micMissing = bind(lib,
		binding{&a.getMicMode, "AVControlCenterMicrophoneModesModuleGetMicrophoneModeForBundleID"},
		binding{&a.setMicMode, "AVControlCenterMicrophoneModesModuleSetMicrophoneModeForBundleID"},
		binding{&a.supportedMicModes, "AVControlCenterMicrophoneModesModuleGetSupportedMicrophoneModesForBundleID"},
	)
	return a, nil
}

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

func (a *avf) eligible() bool {
	return objc.Send[bool](class("AVCaptureDevice"), selIsEligible)
}

func (a *avf) url(app string) string {
	url := a.getBackgroundURL(nsString(app))
	if url == 0 {
		return ""
	}
	return goString(url.Send(selPath))
}

func (a *avf) setURL(path, app string) {
	a.setBackgroundURL(class("NSURL").Send(selFileURLWithPath, nsString(path)), nsString(app))
}

func (a *avf) effectErr(e Effect) error {
	if symbol := a.effectMissing[e]; symbol != "" {
		return missing(symbol)
	}
	return nil
}

func (a *avf) supported(e Effect, app string) bool {
	if a.effectErr(e) != nil {
		return false
	}
	if a.isEffectSupported == nil {
		return true
	}
	return a.isEffectSupported(a.effects[e], nsString(app))
}

func (a *avf) enabled(e Effect, app string) bool {
	return a.isEffectEnabled(a.effects[e], nsString(app))
}

func (a *avf) setEnabled(e Effect, on bool, app string) {
	a.setEffectEnabled(a.effects[e], on, nsString(app))
}

func (a *avf) micErr() error {
	if a.micMissing != "" {
		return missing(a.micMissing)
	}
	return nil
}

func (a *avf) mic(app string) MicMode {
	return MicMode(a.getMicMode(nsString(app)))
}

// micModes lists the modes the app supports. Setting any other mode raises
// an Objective-C exception.
func (a *avf) micModes(app string) []MicMode {
	list := a.supportedMicModes(nsString(app))
	if list == 0 {
		return nil
	}
	var modes []MicMode
	for i := range objc.Send[int](list, selCount) {
		modes = append(modes, MicMode(objc.Send[int](list.Send(selObjectAtIndex, i), selIntegerValue)))
	}
	return modes
}

func (a *avf) setMic(mode MicMode, app string) bool {
	return a.setMicMode(int(mode), nsString(app))
}

func (a *avf) toggled(app string) bool {
	return a.backgroundToggled(nsString(app))
}

func (a *avf) bundleInfo(appPath string) (bundle, bool) {
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
