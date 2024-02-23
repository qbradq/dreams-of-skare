package dosd

import (
	"errors"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/qbradq/dreams-of-skare/internal/util"
	"nhooyr.io/websocket"
)

type websocketService struct {
	l       net.Listener
	mux     http.ServeMux
	once    sync.Once
	wg      *sync.WaitGroup
	cwg     *sync.WaitGroup
	cl      sync.Mutex
	clients map[string]client
	closed  bool
}

func (s *websocketService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *websocketService) Start(wg *sync.WaitGroup) {
	s.wg = wg
	s.cwg = &sync.WaitGroup{}
	s.clients = map[string]client{}
	fsh := http.FileServer(http.Dir(filepath.Join("data", "www")))
	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fsh.ServeHTTP(w, r)
	}))
	s.mux.Handle("/pty", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn := func(err error) {
			cs := websocket.CloseStatus(err)
			if cs != websocket.StatusNormalClosure &&
				cs != websocket.StatusGoingAway {
				log.Println("error:", err)
			}
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			fn(err)
			return
		}
		c := &websocketClient{
			uuid: util.NewUUID(),
			conn: conn,
			ctx:  r.Context(),
		}
		c.DisconnectHook(func() {
			s.cl.Lock()
			delete(s.clients, c.uuid)
			s.cl.Unlock()
		})
		s.cl.Lock()
		if s.closed {
			s.cl.Unlock()
			return
		}
		s.clients[c.uuid] = c
		s.cl.Unlock()
		s.cwg.Add(1)
		c.Start(s.cwg)
		handleClient(c)
		c.Stop()
	}))
	go func() {
		fn := func(err error) {
			if !errors.Is(err, net.ErrClosed) {
				log.Println("error: websocket service exiting with error", err)
			}
			s.Stop()
			gracefulShutdown()
		}
		var err error
		log.Println("info: websocket service starting on 127.0.0.1:80")
		s.l, err = net.Listen("tcp", "127.0.0.1:80")
		if err != nil {
			fn(err)
		}
		if err := http.Serve(s.l, s); err != nil {
			fn(err)
		}
		log.Println("info: websocket service stopped")
	}()
}

func (s *websocketService) Stop() {
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
