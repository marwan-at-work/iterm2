package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"marwan.io/iterm2/api"
)

// New returns a new websocket connection
// that talks to the iTerm2 application.New
// Callers must call the Close() method when done.
// The cookie parameter is optional. If provided,
// it will bypass script authentication prompts.
func New() (*Client, error) {
	h := http.Header{}
	h.Set("origin", "ws://localhost/")
	h.Set("x-iterm2-library-version", "go 3.6")
	h.Set("x-iterm2-disable-auth-ui", "true")
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		h.Set("x-iterm2-cookie", cookie)
	}
	c, resp, err := websocket.DefaultDialer.Dial("ws://localhost:1912", h)
	if err != nil && resp != nil {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error connecting to iTerm2: %v - body: %s", err, b)
	}
	if err != nil {
		return nil, fmt.Errorf("error connecting to iTerm2: %v", err)
	}
	cl := &Client{
		c:    c,
		rpcs: make(map[int64]chan<- *api.ServerOriginatedMessage),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel
	go cl.worker(ctx)
	return cl, nil
}

// Client wraps a websocket client connection to iTerm2.
// Must be instantiated with NewClient.
type Client struct {
	c      *websocket.Conn
	rpcs   map[int64]chan<- *api.ServerOriginatedMessage
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (c *Client) worker(ctx context.Context) {
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
		c.mu.Unlock()
		if !ok {
			fmt.Fprintf(os.Stderr, "could not find call for %d: %v\n", resp.GetId(), &resp)
			continue
		}
		delete(c.rpcs, resp.GetId())
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
	err = c.c.WriteMessage(websocket.BinaryMessage, msg)
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
	c.cancel()
	return c.c.Close()
}

func id(i int64) *int64 {
	return &i
}

func str(s string) *string {
	return &s
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
