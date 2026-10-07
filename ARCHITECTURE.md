# Architecture

macOS keeps each app's video effects and mic mode in the camera daemon,
`cameracaptured`, keyed by bundle identifier. Control Center changes them
through private C functions in
`/System/Library/Frameworks/AVFoundation.framework/AVFoundation`. `onair`
calls the same functions through [purego](https://github.com/ebitengine/purego),
without cgo. The public `AVCaptureDevice` class methods only change the calling
process's own record, so they are no use from a terminal.

## Code map

| Path | Holds |
| --- | --- |
| `cmd/onair/main.go` | CLI: usage, argument parsing, reading apps back after a change, table output, exit codes |
| `internal/controlcenter/controlcenter.go` | `Current` and the setters, which check every app before changing any |
| `internal/controlcenter/effects.go` | `Effect` and `MicMode`, named after the menu labels |
| `internal/controlcenter/apps.go` | App discovery from the Applications folders, `knownApps`, defaults |
| `internal/controlcenter/builtin.go` | Apple's built-in gradient backgrounds |
| `internal/controlcenter/bridge.go` | `bridge` interface: every system call goes through it |
| `internal/controlcenter/objc_darwin.go` | Real `bridge`: binds the private functions with purego |
| `internal/controlcenter/objc_stub.go` | Non-macOS stub, so the module builds and lints on Linux |

## Symbols

These prototypes come from disassembly. There are no headers.

```objc
NSURL    *AVControlCenterVideoEffectsModuleGetBackgroundReplacementURL(NSString *bundleID);
void      AVControlCenterVideoEffectsModuleSetBackgroundReplacementURL(NSURL *url, NSString *bundleID);
BOOL      AVControlCenterVideoEffectsModuleIsEffectEnabledForBundleID(NSString *effect, NSString *bundleID);
void      AVControlCenterVideoEffectsModuleSetEffectEnabledForBundleID(NSString *effect, BOOL on, NSString *bundleID);
BOOL      AVControlCenterVideoEffectsModuleIsEffectSupportedForBundleID(NSString *effect, NSString *bundleID);
BOOL      AVControlCenterVideoEffectsModuleHasBackgroundReplacementBeenToggledForBundleID(NSString *bundleID);
NSInteger AVControlCenterMicrophoneModesModuleGetMicrophoneModeForBundleID(NSString *bundleID);
BOOL      AVControlCenterMicrophoneModesModuleSetMicrophoneModeForBundleID(NSInteger mode, NSString *bundleID);
NSArray  *AVControlCenterMicrophoneModesModuleGetSupportedMicrophoneModesForBundleID(NSString *bundleID);
```

`effect` is an exported `NSString *const`:

| Command | Constant |
| --- | --- |
| `portrait` | `AVControlCenterVideoEffectBackgroundBlur` |
| `studio-light` | `AVControlCenterVideoEffectStudioLighting` |
| `reactions` | `AVControlCenterVideoEffectGestures` |
| `background` | `AVControlCenterVideoEffectBackgroundReplacement` |

Mic modes are the public `AVCaptureMicrophoneMode` values.

## Invariants

- Tests use `fakeBridge` (`useFake`) and never call a setter on the real
  `bridge`, so `go test` never changes your settings.
- The background symbols are required. The rest are optional: when one is
  missing, only the command that needs it fails.
- A mic mode is checked against the app's supported list before it is set.
  Setting an unsupported one raises an Objective-C exception that kills the
  process.

## Gotchas

- The menu's Reactions switch is `...Gestures`. `...Reactions` is a different
  setting the menu does not show.
- Edge Light cannot be switched. The generic functions ignore
  `...RingLight`, and `SetRingLightActiveForBundleID` changes a value but not
  the light. Control Center draws it with its own helper,
  `com.apple.controlcenter.ringlighthelper`, which nothing else can reach.
- Only the mic setter reports failure, so `onair` reads each app back after a
  change and prints that.
- The daemon cannot list apps, so `onair` scans the Applications folders.
  Default apps are `knownApps` plus any whose background was toggled before.
- A path can come back spelled differently, such as with `Versions/A/` added.
