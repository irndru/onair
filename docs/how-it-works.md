# How it works

macOS keeps each app's video effects and mic mode in the camera daemon,
`cameracaptured`, keyed by bundle identifier. Control Center changes them
through private C functions in
`/System/Library/Frameworks/AVFoundation.framework/AVFoundation`. `videobg`
calls the same functions through [purego](https://github.com/ebitengine/purego).
The public `AVCaptureDevice` class methods only change the calling process's
own record, so they are no use from a terminal.

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

The background symbols are required. The rest are optional: when one is
missing, only the command that needs it fails.

## Gotchas

- The menu's Reactions switch is `...Gestures`. `...Reactions` is a different
  setting the menu does not show.
- Edge Light cannot be switched. The generic functions ignore
  `...RingLight`, and `SetRingLightActiveForBundleID` changes a value but not
  the light. Control Center draws it with its own helper,
  `com.apple.controlcenter.ringlighthelper`, which nothing else can reach.
- Setting a mic mode outside the app's supported list raises an Objective-C
  exception that kills the process, so `videobg` checks the list first. Modes
  are the public `AVCaptureMicrophoneMode` values.
- Only the mic setter reports failure. `videobg` reads each app back after a
  change and prints that.
- The daemon cannot list apps, so `videobg` scans the Applications folders.
  Default apps are `knownApps` plus any whose background was toggled before.
- A path can come back spelled differently, such as with `Versions/A/` added.
