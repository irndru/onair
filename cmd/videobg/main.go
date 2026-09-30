// Command videobg sets the macOS camera background per app.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"text/tabwriter"

	"github.com/AndrewMcCraeCA/videobg/internal/videofx"
)

const usage = `usage: videobg <command> [args]

  set <image> [app...]   set the background image and turn the effect on
  on|off [app...]        turn the effect on or off, keeping the image
  status [app...]        show the effect and image for each app
  apps                   list the apps videobg knows about
  backgrounds            list the built-in images
  version                print the version
  help                   print this help

<image> is a built-in name or the path of an image file.
<app> is a name from "videobg apps" or a bundle identifier.
With no app, a command applies to every default app.
`

type usageError string

func (e usageError) Error() string { return string(e) }

func main() {
	err := run(os.Args[1:], os.Stdout)
	var usageErr usageError
	switch {
	case err == nil:
	case errors.As(err, &usageErr):
		if usageErr != "" {
			fmt.Fprintf(os.Stderr, "videobg: %s\n\n", usageErr)
		}
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	default:
		fmt.Fprintln(os.Stderr, "videobg:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return usageError("")
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return nil
	case "version":
		fmt.Fprintln(out, "videobg", version())
		return nil
	case "backgrounds":
		return printBackgrounds(out)
	case "apps":
		apps, err := videofx.Apps()
		if err != nil {
			return err
		}
		return printApps(out, apps)
	case "set":
		if len(args) == 0 {
			return usageError("set needs an image")
		}
		image, err := videofx.ResolveImage(args[0])
		if err != nil {
			return err
		}
		return apply(out, args[1:], func(ids []string) error { return videofx.SetImage(image, ids...) })
	case "on", "off":
		return apply(out, args, func(ids []string) error { return videofx.SetEnabled(cmd == "on", ids...) })
	case "status":
		return apply(out, args, func([]string) error { return nil })
	}
	return usageError(fmt.Sprintf("unknown command %q", cmd))
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return cmp.Or(info.Main.Version, "(devel)")
	}
	return "(devel)"
}

// apply runs change on the selected apps, then prints their state.
func apply(out io.Writer, args []string, change func(ids []string) error) error {
	apps, err := videofx.Apps()
	if err != nil {
		return err
	}
	ids, err := selectApps(apps, args)
	if err != nil {
		return err
	}
	if err := change(ids); err != nil {
		return err
	}
	states := make([]videofx.State, len(ids))
	for i, id := range ids {
		if states[i], err = videofx.Current(id); err != nil {
			return err
		}
	}
	return printStates(out, apps, states)
}

func selectApps(apps []videofx.App, args []string) ([]string, error) {
	if len(args) == 0 {
		ids := videofx.Defaults(apps)
		if len(ids) == 0 {
			return nil, errors.New("no default apps found: name an app or bundle identifier")
		}
		return ids, nil
	}
	var ids []string
	for _, arg := range args {
		id, err := videofx.ResolveApp(apps, arg)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func printStates(out io.Writer, apps []videofx.App, states []videofx.State) error {
	names := map[string]string{}
	for _, a := range apps {
		names[a.BundleID] = a.Name
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, s := range states {
		state := "off"
		if s.Enabled {
			state = "on"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", cmp.Or(names[s.App], s.App), state, cmp.Or(s.Image, "-"))
	}
	return w.Flush()
}

func printApps(out io.Writer, apps []videofx.App) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "APP\tBUNDLE ID\tDEFAULT")
	for _, a := range apps {
		def := "-"
		if a.Default() {
			def = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.Name, a.BundleID, def)
	}
	return w.Flush()
}

func printBackgrounds(out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, name := range videofx.Builtin() {
		path, _ := videofx.BuiltinPath(name)
		fmt.Fprintf(w, "%s\t%s\n", name, path)
	}
	return w.Flush()
}
