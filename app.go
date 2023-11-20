package iterm2

import (
	"fmt"
	"io"

	"marwan.io/iterm2/api"
	"marwan.io/iterm2/client"
)

// App represents an open iTerm2 application
type App interface {
	io.Closer

	CreateWindow() (Window, error)
	ListWindows() ([]Window, error)
	SelectMenuItem(item string) error
	Activate(raiseAllWindows, ignoreOtherApps bool) error
}

// NewApp establishes a connection
// with iTerm2 and returns an App.
// Name is an optional parameter that
// can be used to register your application
// name with iTerm2 so that it doesn't
// require explicit permissions every
// time you run the plugin.
func NewApp(name string) (App, error) {
	c, err := client.New(name)
	if err != nil {
		return nil, err
	}

	return &app{c: c}, nil
}

type app struct {
	c *client.Client
}

func (w *app) Activate(raiseAllWindows bool, ignoreOtherApps bool) error {

	t := true
	_, err := w.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ActivateRequest{ActivateRequest: &api.ActivateRequest{
			OrderWindowFront: &t,
			ActivateApp: &api.ActivateRequest_App{
				RaiseAllWindows:   &raiseAllWindows,
				IgnoringOtherApps: &ignoreOtherApps,
			},
		}},
	})
	return err
}

func (a *app) CreateWindow() (Window, error) {
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: &api.CreateTabRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create window tab: %w", err)
	}
	ctr := resp.GetCreateTabResponse()
	if ctr.GetStatus() != api.CreateTabResponse_OK {
		return nil, fmt.Errorf("unexpected window tab status: %s", ctr.GetStatus())
	}
	return &window{
		c:       a.c,
		id:      ctr.GetWindowId(),
		session: ctr.GetSessionId(),
	}, nil
}

func (a *app) ListWindows() ([]Window, error) {
	list := []Window{}
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &api.ListSessionsRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not list sessions: %w", err)
	}
	for _, w := range resp.GetListSessionsResponse().GetWindows() {
		list = append(list, &window{
			c:  a.c,
			id: w.GetWindowId(),
		})
	}
	return list, nil
}

func (a *app) Close() error {
	return a.c.Close()
}

func str(s string) *string {
	return &s
}

func (a *app) SelectMenuItem(item string) error {
	resp, err := a.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_MenuItemRequest{
			MenuItemRequest: &api.MenuItemRequest{
				Identifier: &item,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error selecting menu item %q: %w", item, err)
	}
	if resp.GetMenuItemResponse().GetStatus() != api.MenuItemResponse_OK {
		return fmt.Errorf("menu item %q returned unexpected status: %q", item, resp.GetMenuItemResponse().GetStatus().String())
	}
	return nil
}
