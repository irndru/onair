package videofx

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

type App struct {
	BundleID string
	Name     string
	Toggled  bool // the background has been switched on or off for this app before
	Known    bool
}

// Default reports whether commands given no app apply to this one.
func (a App) Default() bool { return a.Known || a.Toggled }

// knownApps use the camera without necessarily declaring it in Info.plist.
var knownApps = map[string]bool{
	"com.apple.FaceTime":         true,
	"com.apple.PhotoBooth":       true,
	"com.apple.Safari":           true,
	"com.google.Chrome":          true,
	"com.microsoft.Edge":         true,
	"com.microsoft.teams2":       true,
	"com.tinyspeck.slackmacgap":  true,
	"company.thebrowser.Browser": true,
	"org.mozilla.firefox":        true,
	"us.zoom.xos":                true,
	"com.cisco.webexmeetingsapp": true,
	"com.hnc.Discord":            true,
	"com.loom.desktop":           true,
}

// bundle is what bundleInfo reads from an app's Info.plist.
type bundle struct {
	id, displayName, name string
	camera                bool
}

func appDirs() []string {
	dirs := []string{"/Applications", "/System/Applications"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}
	return dirs
}

// Apps lists the installed apps that can use the camera.
func Apps() ([]App, error) { return appsIn(appDirs()) }

func appsIn(dirs []string) ([]App, error) {
	if err := load(); err != nil {
		return nil, err
	}
	var apps []App
	seen := map[string]bool{}
	for _, dir := range dirs {
		for _, pattern := range []string{"*.app", "*/*.app"} {
			paths, _ := filepath.Glob(filepath.Join(dir, pattern))
			for _, path := range paths {
				b, ok := bundleInfo(path)
				if !ok || seen[b.id] || !(b.camera || knownApps[b.id]) {
					continue
				}
				seen[b.id] = true
				apps = append(apps, App{
					BundleID: b.id,
					Name:     cmp.Or(clean(b.displayName), clean(b.name), strings.TrimSuffix(filepath.Base(path), ".app")),
					Toggled:  toggled(b.id),
					Known:    knownApps[b.id],
				})
			}
		}
	}
	slices.SortFunc(apps, func(a, b App) int { return cmp.Compare(a.BundleID, b.BundleID) })
	return apps, nil
}

// clean drops invisible characters around a name, such as WhatsApp's leading U+200E.
func clean(name string) string {
	return strings.TrimFunc(name, func(r rune) bool { return !unicode.IsGraphic(r) || unicode.IsSpace(r) })
}

// Defaults returns the bundle identifiers of the default apps.
func Defaults(apps []App) []string {
	var ids []string
	for _, a := range apps {
		if a.Default() {
			ids = append(ids, a.BundleID)
		}
	}
	return ids
}

// ResolveApp turns an app name or bundle identifier into a bundle identifier.
func ResolveApp(apps []App, query string) (string, error) {
	for _, a := range apps {
		if strings.EqualFold(query, a.BundleID) || strings.EqualFold(query, a.Name) {
			return a.BundleID, nil
		}
	}
	if strings.Contains(query, ".") {
		return query, nil
	}
	return "", fmt.Errorf("unknown app %q: run \"videobg apps\" for names, or use a bundle identifier", query)
}
