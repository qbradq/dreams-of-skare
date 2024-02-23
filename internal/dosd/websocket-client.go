package dosd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// websocketClient implements client over telnet.
type websocketClient struct {
	uuid  string
	wg    *sync.WaitGroup
	dcFn  func()
	close sync.Once
	in    chan string
	out   chan string
	conn  *websocket.Conn
	ctx   context.Context
}

func (c *websocketClient) Start(wg *sync.WaitGroup) {
	c.wg = wg
	c.in = make(chan string, 16)
	c.out = make(chan string, 128)
	go func() {
		for {
			t, d, err := c.conn.Read(c.ctx)
			if t != websocket.MessageText {
				continue
			}
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("error: websocket connection timeout on read")
				} else {
					log.Println("error: websocket", err)
				}
				break
			}
			c.in <- string(d)
		}
		c.Stop()
		close(c.in)
	}()
	go func() {
		for s := range c.out {
			ctx, cancel := context.WithTimeout(c.ctx, time.Second*15)
			if err := c.conn.Write(ctx, websocket.MessageText, []byte(s)); err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("error: websocket connection timeout on write")
				} else {
					log.Println("error: websocket", err)
				}
				cancel()
				c.Stop()
				return
			}
			cancel()
		}
	}()
}

func (c *websocketClient) Stop() {
	c.close.Do(func() {
		c.conn.Close(websocket.StatusNormalClosure, "disconnecting")
		close(c.out)
		c.wg.Done()
		c.dcFn()
	})
}

func (c *websocketClient) DisconnectHook(fn func()) {
	c.dcFn = fn
}

func (c *websocketClient) GetLine() (string, bool) {
	l, open := <-c.in
	return l, open
}

func (c *websocketClient) PutLine(s string, args ...any) {
	t := fmt.Sprintf(s, args...) + "\r\n"
	select {
	case c.out <- t:
		// Output complete
	default:
		// Channel is flooded
		c.Stop()
	}
}

func (c *websocketClient) PutRaw(s string, args ...any) {
	t := fmt.Sprintf(s, args...)
	select {
	case c.out <- t:
		// Output complete
	default:
		// Channel is flooded
		c.Stop()
	}
}
