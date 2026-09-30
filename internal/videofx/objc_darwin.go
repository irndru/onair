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
	effectBackground  objc.ID
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
)

// load binds the private AVFoundation functions, once.
var load = sync.OnceValue(func() error {
	lib, err := purego.Dlopen(avFoundation, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	missing := func(symbol string) error {
		return fmt.Errorf("this macOS version is not supported: AVFoundation has no %s", symbol)
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
	const effect = "AVControlCenterVideoEffectBackgroundReplacement"
	addr, err := purego.Dlsym(lib, effect)
	if err != nil {
		return missing(effect)
	}
	effectBackground = **(**objc.ID)(unsafe.Pointer(&addr))
	if !objc.Send[bool](class("AVCaptureDevice"), selRespondsToSelector, selIsEligible) {
		return missing("+[AVCaptureDevice isEligibleForBackgroundReplacement]")
	}
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

func isEnabled(bundleID string) bool {
	return effectEnabled(effectBackground, nsString(bundleID))
}

func setEnabled(on bool, bundleID string) {
	setEffectEnabled(effectBackground, on, nsString(bundleID))
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
