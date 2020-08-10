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
}

// NewApp establishes a connection
// with iTerm2 and returns an App
func NewApp() (App, error) {
	c, err := client.New()
	if err != nil {
		return nil, err
	}

	return &app{c: c}, nil
}

type app struct {
	c *client.Client
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
	fmt.Println(resp.GetListSessionsResponse().GetBuriedSessions())
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
