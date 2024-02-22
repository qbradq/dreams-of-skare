package dosd

import (
	"sync"
	"time"
)

// dreamCommand encapsulates a command line from the player and that player's
// account.
type dreamCommand struct {
	Line    string   // Command line
	Account *account // Account associated with the command
}

// dreamCommands is the queue of commands for the dream to execute.
var dreamCommands = make(chan *dreamCommand, 1024*16)

// dreamService is the main game daemon loop.
func dreamService(wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second / 10)
mainLoop:
	for {
		select {
		case <-ticker.C:
		case c := <-dreamCommands:
			if c == nil {
				break mainLoop
			}
			if c.Account.Character.Player.Client == nil {
				continue
			}
			executeCommand(c.Line, c.Account)
		case <-done:
			break mainLoop
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}
