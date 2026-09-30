# Internals

## Where the setting lives

macOS stores the Background setting per app, keyed by bundle identifier,
inside the camera daemon `cameracaptured`. There is no file to read and no way
to list the apps it has a record for.

`AVCaptureDevice` has class methods such as `+setBackgroundReplacementURL:`,
but they only change the record of the calling process. Called from a
terminal, they change nothing useful.

## The five functions

Control Center changes other apps' records through C functions exported by
`/System/Library/Frameworks/AVFoundation.framework/AVFoundation`. They have no
headers. These prototypes come from disassembly:

```objc
NSURL *AVControlCenterVideoEffectsModuleGetBackgroundReplacementURL(NSString *bundleID);
void   AVControlCenterVideoEffectsModuleSetBackgroundReplacementURL(NSURL *url, NSString *bundleID);
BOOL   AVControlCenterVideoEffectsModuleIsEffectEnabledForBundleID(NSString *effect, NSString *bundleID);
void   AVControlCenterVideoEffectsModuleSetEffectEnabledForBundleID(NSString *effect, BOOL on, NSString *bundleID);
BOOL   AVControlCenterVideoEffectsModuleHasBackgroundReplacementBeenToggledForBundleID(NSString *bundleID);
```

`effect` is the exported constant
`AVControlCenterVideoEffectBackgroundReplacement`, an `NSString *const`.

`+[AVCaptureDevice isEligibleForBackgroundReplacement]` reports whether the
Mac supports the effect. `videobg` checks it before any change.

The set functions persist before they return. They need no admin rights,
entitlements or privacy prompts. The effect applies to the built-in camera
only, not USB webcams.

The daemon may hand back a different spelling of the path it was given. A
built-in gradient comes back with `Versions/A/` in it.

## Calling them from Go

`videobg` uses [purego](https://github.com/ebitengine/purego) rather than cgo,
so it builds with `CGO_ENABLED=0` and cross-vets from Linux.

- `load()` opens AVFoundation once and binds each function with
  `purego.Dlsym` and `purego.RegisterFunc`. `purego.RegisterLibFunc` would
  panic on a missing symbol. `load()` instead returns an error naming it.
- `Dlsym` on the effect constant gives the address of the pointer, so the code
  dereferences it once to get the `NSString`.
- `objc.ID` is pointer-sized and passes as a pointer argument.
- purego converts a Go `string` argument to a C string and copies a returned
  C string back, so `NSString` conversion is `stringWithUTF8String:` and
  `UTF8String`.
- The `objc` package loads libobjc itself. AVFoundation pulls in Foundation.

Objects are autoreleased and never drained. The process exits straight away.

## App discovery

Because the daemon cannot list apps, `videobg` looks for them on disk. It
scans `/Applications`, `/System/Applications` and `~/Applications`, one level
of subdirectory deep, and reads each `Contents/Info.plist` with
`NSDictionary`. It keeps an app when either:

- its `Info.plist` has `NSCameraUsageDescription`, or
- its bundle identifier is on the known list.

The known list exists because some apps use the camera without that key.
Chrome is one. The list:

```
com.apple.FaceTime          com.apple.PhotoBooth     com.apple.Safari
com.google.Chrome           com.microsoft.Edge       com.microsoft.teams2
com.tinyspeck.slackmacgap   company.thebrowser.Browser
org.mozilla.firefox         us.zoom.xos              com.cisco.webexmeetingsapp
com.hnc.Discord             com.loom.desktop
```

An app is a default target when it is on the known list or
`HasBackgroundReplacementBeenToggledForBundleID` is true for it. This stops a
bare `videobg set` from writing records for every app that merely declares
camera use.

The name is `CFBundleDisplayName`, then `CFBundleName`, then the `.app` file
name, with invisible characters trimmed.

## Built-in gradients

Apple ships the gradients Control Center offers in:

```
/System/Library/PrivateFrameworks/Portrait.framework/Resources/backgrounds/gradient/
```

as `1_blue.png` to `9_dark.png`. `videobg` maps a name to each file.

## Why no animation

The effect takes a still image. Given an animated GIF it shows the first
frame. `videobg` does not try to work around this.
