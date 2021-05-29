// Package scaffold lets you declare a preset
// window with tabs, pains, env vars, and other
// configurations to get your multi-terminal program
// up fast.
//
// This Package is EXPERIMENTAL and its APIs are likely
// to change before becoming stable.
package scaffold

import (
	"fmt"

	"golang.org/x/sync/errgroup"
	"marwan.io/iterm2"
)

// WindowSpec specifies a window configuration.
type WindowSpec struct {
	Title string
	Tabs  []TabSpec
}

// Run takes a window spec and creates a new iTerm2 session and uses it
// to create a new window with the given specs.
func Run(appName string, w WindowSpec) error {
	if w.Title == "" {
		return fmt.Errorf("window must have a title")
	}
	if len(w.Tabs) == 0 {
		return fmt.Errorf("window must have at least 1 tab")
	}
	app, err := iterm2.NewApp(appName)
	if err != nil {
		return fmt.Errorf("iterm2.NewApp: %w", err)
	}
	defer app.Close()
	window, err := app.CreateWindow()
	if err != nil {
		return fmt.Errorf("app.CreateWindow: %w", err)
	}
	err = window.SetTitle(w.Title)
	if err != nil {
		return fmt.Errorf("window.SetTitle: %w", err)
	}
	var eg errgroup.Group
	for i, ts := range w.Tabs {
		var tab iterm2.Tab
		if i == 0 {
			tabs, err := window.ListTabs()
			if err != nil {
				return fmt.Errorf("window.ListTabs: %w", err)
			}
			tab = tabs[0]
		} else {
			tab, err = window.CreateTab()
			if err != nil {
				return fmt.Errorf("window.CreateTab: %w", err)
			}
		}
		ts := ts
		eg.Go(func() error { return createTab(tab, ts) })
	}
	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("createTab: %w", err)
	}
	windowTabs, err := window.ListTabs()
	if err != nil {
		return err
	}
	ss, err := windowTabs[0].ListSessions()
	if err != nil {
		return fmt.Errorf("first tab sessions: %w", err)
	}
	err = ss[0].Activate(true, true)
	if err != nil {
		return fmt.Errorf("session.Activate: %w", err)
	}
	return nil
}

func createTab(tab iterm2.Tab, ts TabSpec) error {
	err := tab.SetTitle(ts.Title)
	if err != nil {
		return fmt.Errorf("tab.SetTitle: %w", err)
	}
	sessions, err := tab.ListSessions()
	if err != nil {
		return fmt.Errorf("tab.ListSessions: %w", err)
	}
	sesh := sessions[0]
	if ts.Dir != "" {
		err = sesh.SendText(fmt.Sprintf("cd %v\n", ts.Dir))
		if err != nil {
			return fmt.Errorf("error changing directory: %w", err)
		}
	}
	if ts.Env != nil {
		for _, e := range ts.Env.GetEnv() {
			err = sesh.SendText(fmt.Sprintf("export %s\n", e))
			if err != nil {
				return fmt.Errorf("error exporting env: %w", err)
			}
		}
	}
	if ts.OnCreate != nil {
		err = ts.OnCreate(sesh)
		if err != nil {
			return err
		}
	}
	if ts.Pane != nil {
		pane, err := sesh.SplitPane(iterm2.SplitPaneOptions{
			Vertical: true,
		})
		if err != nil {
			return fmt.Errorf("sesh.SplitPane: %w", err)
		}
		if ts.Pane.OnCreate != nil {
			err = ts.Pane.OnCreate(pane)
			if err != nil {
				return fmt.Errorf("pane.OnCreate: %w", err)
			}
		}
	}
	return nil
}

// EnvGetter defines an interface that can be used
// to retrieve environment variables to be set on
// an iTerm2 session. The reason for this interface
// is that so you can specify dynamic tabs such as tabs
// being read from a file instead of having to statically
// create them. If you want to statically pass a set of env vars to
// a TabSpec, just use scaffold.Env which is a slice of strings that
// implements this interface
type EnvGetter interface {
	GetEnv() []string
}

// Env is a slice of environment variables
// to be exported in an iTerm2 session
type Env []string

// GetEnv implements EnvGetter
func (e Env) GetEnv() []string {
	return []string(e)
}

// TabSpec specifies a main tab in your iTerm2 window
type TabSpec struct {
	Title    string
	Dir      string
	Env      EnvGetter
	OnCreate func(s iterm2.Session) error
	Pane     *PaneSpec
}

// PaneSpec specifies a vertical right pane within a tab
type PaneSpec struct {
	OnCreate func(s iterm2.Session) error
}
