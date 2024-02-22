package dosd

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/qbradq/dreams-of-skare/internal/game"
)

var banner []string

var dream *game.Dream

var sshSrv = &sshService{}

var tnSrv = &telnetService{}

var shutdownOnce sync.Once

var dreamDoneChan = make(chan struct{}, 1)

func Main() {
	// Load global resources
	bd, err := os.ReadFile(filepath.Join("data", "banner.txt"))
	if err != nil {
		log.Fatalf("fatal: %v", err)
	}
	banner = strings.Split(string(bd), "\n")
	// Load accounts
	d, err := os.ReadFile(filepath.Join("saves", "accounts.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("fatal: %v", err)
		}
	} else if err := json.Unmarshal(d, &accounts); err != nil {
		log.Fatalf("fatal: %v", err)
	}
	// Load dream
	d, err = os.ReadFile(filepath.Join("saves", "dream.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("fatal: %v", err)
		}
		dream = game.NewDream()
	} else {
		dream = &game.Dream{}
		if err := json.Unmarshal(d, dream); err != nil {
			log.Fatalf("fatal: %v", err)
		}
	}
	// Start services
	wg := &sync.WaitGroup{}
	wg.Add(1)
	sshSrv.Start(wg)
	wg.Add(1)
	tnSrv.Start(wg)
	wg.Add(1)
	go dreamService(wg, dreamDoneChan)
	// Trap signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		gracefulShutdown()
	}()
	// Wait for services to quit
	log.Println("info: dosd running")
	wg.Wait()
	log.Println("info: dosd exiting normally")
}

func gracefulShutdown() {
	shutdownOnce.Do(func() {
		sshSrv.Stop()
		tnSrv.Stop()
		close(dreamDoneChan)
	})
}

func handleClient(c client) {
	// Welcome banner
	for _, l := range banner {
		c.PutLine(l)
	}
	// Account login
	a := accountLogin(c)
	if a == nil {
		return
	}
	defer func() { a.Character.Player.Client = nil }()
	// Insert character into the dream
	if !dream.InsertPlayer(a.Character) {
		return
	}
	// Input loop
	for {
		l, ok := c.GetLine()
		if !ok {
			break
		}
		dreamCommands <- &dreamCommand{
			Line:    l,
			Account: a,
		}
	}
}
