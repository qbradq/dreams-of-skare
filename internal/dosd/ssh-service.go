package dosd

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/gliderlabs/ssh"
	"github.com/qbradq/dreams-of-skare/internal/util"
	"golang.org/x/term"
)

type sshService struct {
	once    sync.Once
	wg      *sync.WaitGroup
	cwg     *sync.WaitGroup
	cl      sync.Mutex
	clients map[string]client
	closed  bool
	l       net.Listener
}

func (s *sshService) Start(wg *sync.WaitGroup) {
	s.wg = wg
	s.cwg = &sync.WaitGroup{}
	s.clients = map[string]client{}
	go func() {
		ssh.Handle(func(ss ssh.Session) {
			c := &sshClient{
				uuid: util.NewUUID(),
				s:    ss,
				term: term.NewTerminal(ss, ""),
			}
			s.cl.Lock()
			if s.closed {
				s.cl.Unlock()
				return
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
		})
		var err error
		s.l, err = net.Listen("tcp", "127.0.0.1:22")
		if err != nil {
			log.Println("error: ", err)
			gracefulShutdown()
		}
		log.Println("info: ssh service starting on 127.0.0.1:22")
		if err := ssh.Serve(s.l, nil); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("error: ssh service closing with error %v", err)
			}
		}
		log.Println("info: ssh service stopped")
		s.Stop()
		gracefulShutdown()
	}()
}

func (s *sshService) Stop() {
	s.once.Do(func() {
		if s.l != nil {
			s.l.Close()
		}
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
