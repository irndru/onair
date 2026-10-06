# onair

Set the macOS video effects and mic mode from the command line. `onair`
changes what Control Center > Video Effects shows for each video call app:
Portrait, Studio Light, Reactions, Background and Mic Mode. You can script it
or put it on a schedule. Edge Light is not supported: see
[docs/how-it-works.md](docs/how-it-works.md#gotchas).

```sh
onair status
onair portrait on                       # blur the background
onair studio-light on facetime
onair background blue                   # a built-in gradient, for every default app
onair background ~/Pictures/office.jpg  # your own image
onair background dark zoom.us "Google Chrome"
onair background off                    # turn the background off, keep the image
onair mic-mode voice-isolation
```

## Install

```sh
GOPRIVATE=github.com/irndru/* go install github.com/irndru/onair/cmd/onair@latest
```

The repository is private, so git needs access to it. Or clone it and run
`make install`.

## Requirements

- A Mac with Apple silicon
- macOS 15 or later
- The built-in camera. USB webcams do not get the effect.

Effects other than Background need a macOS version that has them. On an older
one, those commands fail and name the missing symbol; the rest still work.

No admin rights or privacy prompts are needed.

## Commands

| Command | Does |
| --- | --- |
| `onair status [app...]` | Show every effect, the image and the mic mode for each app |
| `onair apps` | List the apps onair knows about |
| `onair backgrounds` | List the built-in images |
| `onair version` | Print the version |
| `onair help` | Print usage |

The rest follow the Video Effects menu, in its order:

| Command | Does |
| --- | --- |
| `onair portrait on\|off [app...]` | Turn Portrait on or off |
| `onair studio-light on\|off [app...]` | Turn Studio Light on or off |
| `onair edge-light` | Fails: Edge Light can only be switched in Control Center |
| `onair reactions on\|off [app...]` | Turn Reactions on or off |
| `onair background on\|off [app...]` | Turn Background on or off, keeping the image |
| `onair background <image> [app...]` | Set the background image and turn it on |
| `onair mic-mode <mode> [app...]` | Set the mic mode |

Each command and mode is the menu label in lower case, with hyphens for
spaces. `<mode>` is `standard`, `voice-isolation` or `wide-spectrum`. Not every
app supports every mode. A command that names a mode or effect an app does
not support fails and changes nothing.

## Images

`<image>` is either a built-in name or the path of an image file.

`onair backgrounds` lists the built-in names: `blue`, `purple`, `pink`, `red`,
`orange`, `yellow`, `green`, `off-white` and `dark`. These are Apple's own
gradients.

A built-in name wins over a file with the same name, and `on` and `off` win
over files named `on` and `off`. Use `./blue` to mean the file.

Animated images do not animate. A GIF shows its first frame.

## Apps

macOS keeps a separate background for each app. `onair apps` lists the
installed apps that can use the camera:

```
APP            BUNDLE ID             DEFAULT
FaceTime       com.apple.FaceTime    yes
Photo Booth    com.apple.PhotoBooth  yes
Notes          com.example.Notes     -
```

With no `[app...]`, a command applies to every app marked `yes`. An app is a
default when it is a well-known video call app or browser, or when its
background has been switched on or off before.

Name an app by its `APP` name (case does not matter, but the whole name must
match) or by its bundle identifier. A bundle identifier works even when the
app is not in the list. To find one:

```sh
osascript -e 'id of app "Zoom"'
```

## Output and exit codes

Every command that takes `[app...]` prints a header and one line per app:
each effect as `on` or `off`, the image path and the mic mode. `-` means the
app does not support it, or has no image.

```
APP            PORTRAIT  STUDIO-LIGHT  REACTIONS  BACKGROUND  IMAGE                          MIC-MODE
Photo Booth    off       on            on         on          /Users/me/Pictures/office.jpg  voice-isolation
Google Chrome  off       off           on         off         -                              -
```

| Exit code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | The command failed. The reason is on stderr. |
| 2 | The command line was wrong. Usage is on stderr. |

For debug output on stderr, put `-v` before the command or set
`ONAIR_DEBUG=1`:

```sh
onair -v portrait on zoom.us
```

## How it works

`onair` calls the same functions in AVFoundation that Control Center calls.
They are private: Apple does not document them and may change or remove them
in any macOS update. If that happens, `onair` reports the missing symbol and
changes nothing.

See [docs/how-it-works.md](docs/how-it-works.md) for the details and
[docs/development.md](docs/development.md) to build and test it.

## Licence

[MIT](LICENSE)
