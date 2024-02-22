package dosd

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/qbradq/dreams-of-skare/internal/util"
)

// telnetService implements a client service over telnet.
type telnetService struct {
	l       net.Listener
	once    sync.Once
	wg      *sync.WaitGroup
	cwg     *sync.WaitGroup
	cl      sync.Mutex
	clients map[string]client
	closed  bool
}

func (s *telnetService) Start(wg *sync.WaitGroup) {
	s.wg = wg
	s.cwg = &sync.WaitGroup{}
	s.clients = map[string]client{}
	go func() {
		fn := func(err error) {
			if !errors.Is(err, net.ErrClosed) {
				log.Println("error: telnet service exiting with error", err)
			}
			s.Stop()
			gracefulShutdown()
		}
		var err error
		log.Println("info: telnet service start on 127.0.0.1:23")
		s.l, err = net.Listen("tcp", "127.0.0.1:23")
		if err != nil {
			fn(err)
		}
		for {
			conn, err := s.l.Accept()
			if err != nil {
				fn(err)
				break
			}
			c := &telnetClient{
				uuid: util.NewUUID(),
				conn: conn,
			}
			s.cl.Lock()
			if s.closed {
				s.cl.Unlock()
				break
			}
			s.clients[c.uuid] = c
			s.cl.Unlock()
			c.DisconnectHook(func() {
				s.cl.Lock()
				delete(s.clients, c.uuid)
				s.cl.Unlock()
			})
			s.cwg.Add(1)
			c.Start(s.cwg)
			handleClient(c)
			c.Stop()
		}
		log.Println("info: telnet service stopped")
	}()
}

func (s *telnetService) Stop() {
	s.once.Do(func() {
		s.l.Close()
		s.cl.Lock()
		clients := make([]client, 0, len(s.clients))
		for _, c := range s.clients {
			clients = append(clients, c)
		}
		s.cl.Unlock()
		for _, c := range clients {
			c.Stop()
		}
		s.cl.Lock()
		s.closed = true
		s.cl.Unlock()
		s.cwg.Wait()
		s.wg.Done()
	})
}
