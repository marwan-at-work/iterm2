package iterm2

import (
	"fmt"

	"marwan.io/iterm2/api"
	"marwan.io/iterm2/client"
)

// Session represents an iTerm2 Session which is a pane
// within a Tab where the terminal is active
type Session interface {
	SendText(s string) error
	Activate(selectTab, orderWindowFront bool) error
	SplitPane(opts SplitPaneOptions) (Session, error)
	GetSessionID() string
}

// SplitPaneOptions for customizing the new pane session.
// More options can be added here as needed
type SplitPaneOptions struct {
	Vertical bool
}

type session struct {
	c  *client.Client
	id string
}

func (s *session) SendText(t string) error {
	resp, err := s.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_SendTextRequest{
			SendTextRequest: &api.SendTextRequest{
				Session: &s.id,
				Text:    &t,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error sending text to session %q: %w", s.id, err)
	}
	if status := resp.GetSendTextResponse().GetStatus(); status != api.SendTextResponse_OK {
		return fmt.Errorf("unexpected status for session %q: %s", s.id, status)
	}
	return nil
}

func (s *session) Activate(selectTab, orderWindowFront bool) error {
	resp, err := s.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &api.ActivateRequest{
				Identifier: &api.ActivateRequest_SessionId{
					SessionId: s.id,
				},
				SelectTab:        &selectTab,
				OrderWindowFront: &orderWindowFront,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error activating session %q: %w", s.id, err)
	}
	if status := resp.GetActivateResponse().GetStatus(); status != api.ActivateResponse_OK {
		return fmt.Errorf("unexpected status for activate request: %s", status)
	}
	return nil
}

func (s *session) SplitPane(opts SplitPaneOptions) (Session, error) {
	direction := api.SplitPaneRequest_HORIZONTAL.Enum()
	if opts.Vertical {
		direction = api.SplitPaneRequest_VERTICAL.Enum()
	}
	resp, err := s.c.Call(&api.ClientOriginatedMessage{
		Submessage: &api.ClientOriginatedMessage_SplitPaneRequest{
			SplitPaneRequest: &api.SplitPaneRequest{
				Session:        &s.id,
				SplitDirection: direction,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error splitting pane: %w", err)
	}
	spResp := resp.GetSplitPaneResponse()
	if len(spResp.GetSessionId()) < 1 {
		return nil, fmt.Errorf("expected at least one new session in split pane")
	}
	return &session{
		c:  s.c,
		id: spResp.GetSessionId()[0],
	}, nil
}

func (s *session) GetSessionID() string {
	return s.id
}
