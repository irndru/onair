# videobg

Set the macOS video effects from the command line. `videobg` changes what
Control Center > Video Effects shows for each video call app: Background,
Portrait, Studio Light, Edge Light, Reactions and Mic Mode. You can script it
or put it on a schedule.

```sh
videobg set blue                     # a built-in gradient, for every default app
videobg set ~/Pictures/office.jpg    # your own image
videobg set dark zoom.us "Google Chrome"
videobg off                          # turn the background off, keep the image
videobg portrait on                  # blur the background
videobg studio on facetime
videobg mic isolation
videobg status
```

## Install

```sh
go install github.com/AndrewMcCraeCA/videobg/cmd/videobg@latest
```

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
| `videobg set <image> [app...]` | Set the background image and turn the effect on |
| `videobg on [app...]` | Turn the background on, keeping the image |
| `videobg off [app...]` | Turn the background off, keeping the image |
| `videobg <effect> on\|off [app...]` | Turn an effect on or off |
| `videobg mic <mode> [app...]` | Set the mic mode |
| `videobg status [app...]` | Show every effect, the mic mode and the image for each app |
| `videobg apps` | List the apps videobg knows about |
| `videobg backgrounds` | List the built-in images |
| `videobg version` | Print the version |
| `videobg help` | Print usage |

## Effects and mic mode

| `<effect>` | Control Center name |
| --- | --- |
| `background` | Background |
| `portrait` | Portrait |
| `studio` | Studio Light |
| `edge` | Edge Light |
| `reactions` | Reactions |

`videobg background on` is the same as `videobg on`.

`<mode>` is `standard`, `isolation` (Voice Isolation) or `wide` (Wide
Spectrum). Not every app supports every mode. A command that names a mode or
effect an app does not support fails and changes nothing.

## Images

`<image>` is either a built-in name or the path of an image file.

`videobg backgrounds` lists the built-in names: `blue`, `purple`, `pink`, `red`,
`orange`, `yellow`, `green`, `off-white` and `dark`. These are Apple's own
gradients.

A built-in name wins over a file with the same name. Use `./blue` to mean the
file.

Animated images do not animate. A GIF shows its first frame.

## Apps

macOS keeps a separate background for each app. `videobg apps` lists the
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
each effect as `on` or `off`, the mic mode and the image path. `-` means the
app does not support it, or has no image.

```
APP            BACKGROUND  PORTRAIT  STUDIO  EDGE  REACTIONS  MIC        IMAGE
Photo Booth    on          off       on      off   on         isolation  /Users/me/Pictures/office.jpg
Google Chrome  off         off       off     off   on         -          -
```

| Exit code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | The command failed. The reason is on stderr. |
| 2 | The command line was wrong. Usage is on stderr. |

## How it works

`videobg` calls the same functions in AVFoundation that Control Center calls.
They are private: Apple does not document them and may change or remove them
in any macOS update. If that happens, `videobg` reports the missing symbol and
changes nothing.

See [docs/internals.md](docs/internals.md) for the details and
[docs/development.md](docs/development.md) to build and test it.

## Licence

[MIT](LICENSE)
