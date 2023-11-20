package iterm2

import (
	"fmt"
	"strconv"

	"marwan.io/iterm2/api"
	"marwan.io/iterm2/client"
)

// Window represents an iTerm2 Window
type Window interface {
	SetTitle(s string) error
	CreateTab() (Tab, error)
	ListTabs() ([]Tab, error)
	Activate() error
}

type window struct {
	c       *client.Client
	id      string
	session string
}

func (w *window) CreateTab() (Tab, error) {
	resp, err := w.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: &api.CreateTabRequest{
				WindowId: str(w.id),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create tab for window %q: %w", w.id, err)
	}
	ctr := resp.GetCreateTabResponse()
	if ctr.GetStatus() != api.CreateTabResponse_OK {
		return nil, fmt.Errorf("unexpected tab status: %s", ctr.GetStatus())
	}
	return &tab{
		c:        w.c,
		id:       strconv.Itoa(int(ctr.GetTabId())),
		windowID: w.id,
	}, nil
}

func (w *window) ListTabs() ([]Tab, error) {
	list := []Tab{}
	resp, err := w.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &api.ListSessionsRequest{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not list sessions: %w", err)
	}
	for _, window := range resp.GetListSessionsResponse().GetWindows() {
		if window.GetWindowId() != w.id {
			continue
		}
		for _, t := range window.GetTabs() {
			list = append(list, &tab{
				c:        w.c,
				id:       t.GetTabId(),
				windowID: w.id,
			})
		}
	}
	return list, nil
}

func (w *window) SetTitle(s string) error {
	_, err := w.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_InvokeFunctionRequest{
			InvokeFunctionRequest: &api.InvokeFunctionRequest{
				Invocation: str(fmt.Sprintf(`iterm2.set_title(title: "%s")`, s)),
				Context: &api.InvokeFunctionRequest_Method_{
					Method: &api.InvokeFunctionRequest_Method{
						Receiver: &w.id,
					},
				},
			},
		},
	})
	return err
}

func (w *window) Activate() error {
	_, err := w.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ActivateRequest{ActivateRequest: &api.ActivateRequest{
			Identifier:       &api.ActivateRequest_WindowId{WindowId: w.id},
			OrderWindowFront: b(true),
		}},
	})
	return err
}
