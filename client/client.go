package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andybrewer/mack"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"marwan.io/iterm2/api"
)

// New returns a new websocket connection that talks to the iTerm2
// application.New Callers must call the Close() method when done. The cookie
// parameter is optional. If provided, it will bypass script authentication
// prompts.
func New(appName string) (*Client, error) {
	// ITERM2_COOKIE is an an environment variable that's set on each terminal
	// session. But it only seems to work the first time, then it gets
	// invalidated. Therefore, we keep trying until it returns an error, then we
	// try to generate a new cookie instead. See
	// https://github.com/marwan-at-work/iterm2/issues/4
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		client, err := newClient(appName, cookie)
		if err == nil {
			return client, nil
		}
	}
	client, err := newClient(appName, "")
	if err != nil {
		return nil, err
	}
	return client, err
}

func newClient(appName, cookie string) (*Client, error) {
	h := http.Header{}
	h.Set("origin", "ws://localhost/")
	h.Set("x-iterm2-library-version", "go 3.6")
	h.Set("x-iterm2-disable-auth-ui", "true")
	// Disable using env var cookie due to
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		h.Set("x-iterm2-cookie", cookie)
	} else {
		resp, err := mack.Tell("iTerm2", fmt.Sprintf("request cookie and key for app named %q", appName))
		if err != nil {
			return nil, fmt.Errorf("AppleScript/tell: %w", err)
		}
		fields := strings.Fields(resp)
		if len(fields) != 2 {
			return nil, fmt.Errorf("incorrect field format: %q", resp)
		}
		h.Set("x-iterm2-cookie", fields[0])
		h.Set("x-iterm2-key", fields[1])
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("os.UserHomeDir: %w", err)
	}
	d := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", filepath.Join(homeDir, "/Library/Application Support/iTerm2/private/socket"))
		},
		HandshakeTimeout: 45 * time.Second,
		Subprotocols:     []string{"api.iterm2.com"},
	}
	c, resp, err := d.Dial("ws://localhost", h)
	if err != nil && resp != nil {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error connecting to iTerm2: %v - body: %s", err, b)
	}
	if err != nil {
		return nil, fmt.Errorf("error connecting to iTerm2: %v", err)
	}
	cl := &Client{
		c:       c,
		rpcs:    make(map[int64]chan<- *api.ServerOriginatedMessage),
		writeCh: make(chan writeReq),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel
	go cl.readWorker(ctx)
	go cl.writeWorker()
	return cl, nil
}

// Client wraps a websocket client connection to iTerm2.
// Must be instantiated with NewClient.
type Client struct {
	c       *websocket.Conn
	rpcs    map[int64]chan<- *api.ServerOriginatedMessage
	mu      sync.Mutex
	cancel  context.CancelFunc
	writeCh chan writeReq
}

type writeReq struct {
	msg  []byte
	resp chan error
}

func (c *Client) writeWorker() {
	for req := range c.writeCh {
		err := c.c.WriteMessage(websocket.BinaryMessage, req.msg)
		req.resp <- err
	}
}

func (c *Client) readWorker(ctx context.Context) {
	for {
		_, msg, err := c.c.ReadMessage()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		var resp api.ServerOriginatedMessage
		err = proto.Unmarshal(msg, &resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		c.mu.Lock()
		ch, ok := c.rpcs[resp.GetId()]
		delete(c.rpcs, resp.GetId())
		c.mu.Unlock()
		if !ok {
			fmt.Fprintf(os.Stderr, "could not find call for %d: %v\n", resp.GetId(), &resp)
			continue
		}
		ch <- &resp
	}
}

// Call sends a request to the iTerm2 server
func (c *Client) Call(req *api.ClientOriginatedMessage) (*api.ServerOriginatedMessage, error) {
	req.Id = id(rand.Int63())
	ch := make(chan *api.ServerOriginatedMessage, 1)
	c.mu.Lock()
	c.rpcs[req.GetId()] = ch
	c.mu.Unlock()
	msg, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	wr := writeReq{msg: msg, resp: make(chan error, 1)}
	c.writeCh <- wr
	err = <-wr.resp
	if err != nil {
		return nil, fmt.Errorf("error writing to websocket: %w", err)
	}
	resp := <-ch
	if resp.GetError() != "" {
		return nil, fmt.Errorf("error from server: %v", resp.GetError())
	}
	return resp, nil
}

// Close closes the websocket connection
// and frees any goroutine resources
func (c *Client) Close() error {
	// TODO: if a *Client.Call is in flight, this will cause it to panic
	close(c.writeCh)
	c.cancel()
	return c.c.Close()
}

func id(i int64) *int64 {
	return &i
}
