package dosd

import (
	"log"
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
		log.Println("info: ssh service listening at 127.0.0.1:22")
		err := ssh.ListenAndServe("127.0.0.1:22", nil)
		s.Stop()
		log.Printf("error: ssh service closing with error %v", err)
	}()
}

func (s *sshService) Stop() {
	s.once.Do(func() {
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
