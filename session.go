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
