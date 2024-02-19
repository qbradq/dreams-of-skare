package dosd

import "sync"

// client is implemented by all player clients.
type client interface {
	// Start starts the client interface running.
	Start(wg *sync.WaitGroup)
	// Stop stops the client running.
	Stop()
	// DisconnectHook sets the disconnection hook for the service.
	DisconnectHook(func())
	// GetLine returns the next non-blank line of text from the player. The
	// second return value is false when the connection is closed.
	GetLine() (string, bool)
	// Put prints a string to the client.
	Put(string, ...any)
}
