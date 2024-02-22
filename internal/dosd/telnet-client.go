package dosd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

// telnetClient implements client over telnet.
type telnetClient struct {
	uuid  string
	wg    *sync.WaitGroup
	dcFn  func()
	close sync.Once
	in    chan string
	out   sync.Mutex
	conn  net.Conn
}

func (c *telnetClient) Start(wg *sync.WaitGroup) {
	c.wg = wg
	c.in = make(chan string, 16)
	go func() {
		s := bufio.NewScanner(c.conn)
		for s.Scan() {
			d := []byte(s.Text())
			l := make([]byte, 0, len(d))
			i := 0
			for {
				if i >= len(d) {
					break
				}
				if d[i] == 0xFF {
					i += 3
					continue
				}
				l = append(l, d[i])
				i++
			}
			if len(l) > 0 {
				c.in <- string(l)
			}
		}
		if err := s.Err(); err != nil {
			log.Println("trace: telnet client disconnected due to error", err)
		}
		c.Stop()
		close(c.in)
	}()
}

func (c *telnetClient) Stop() {
	c.close.Do(func() {
		c.conn.Close()
		c.wg.Done()
		c.dcFn()
	})
}

func (c *telnetClient) DisconnectHook(fn func()) {
	c.dcFn = fn
}

func (c *telnetClient) GetLine() (string, bool) {
	l, open := <-c.in
	return l, open
}

func (c *telnetClient) PutLine(s string, args ...any) {
	t := fmt.Sprintf(s, args...) + "\r\n"
	c.out.Lock()
	c.conn.Write([]byte(t))
	c.out.Unlock()
}

func (c *telnetClient) PutRaw(s string, args ...any) {
	t := fmt.Sprintf(s, args...)
	c.out.Lock()
	c.conn.Write([]byte(t))
	c.out.Unlock()
}
