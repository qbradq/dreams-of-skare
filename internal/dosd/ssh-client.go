package dosd

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

// sshClient implements client over ssh.
type sshClient struct {
	uuid  string
	wg    *sync.WaitGroup
	dcFn  func()
	close sync.Once
	s     ssh.Session
	term  *term.Terminal
	in    chan string
	out   sync.Mutex
}

func (c *sshClient) Start(wg *sync.WaitGroup) {
	c.wg = wg
	c.in = make(chan string, 16)
	go func() {
		for {
			l, err := c.term.ReadLine()
			if err != nil {
				if err != io.EOF {
					log.Printf("error: %v", err)
				}
				c.Stop()
				break
			}
			if l != "" {
				c.in <- l
			}
		}
	}()
}

func (c *sshClient) Stop() {
	c.close.Do(func() {
		c.wg.Done()
		c.s.Close()
		c.dcFn()
	})
}

func (c *sshClient) DisconnectHook(fn func()) {
	c.dcFn = fn
}

func (c *sshClient) GetLine() (string, bool) {
	l, open := <-c.in
	return l, open
}

func (c *sshClient) PutLine(s string, args ...any) {
	t := fmt.Sprintf(s, args...) + "\n"
	c.out.Lock()
	c.term.Write([]byte(t))
	c.out.Unlock()
}

func (c *sshClient) PutRaw(s string, args ...any) {
	t := fmt.Sprintf(s, args...)
	c.out.Lock()
	c.term.Write([]byte(t))
	c.out.Unlock()
}
