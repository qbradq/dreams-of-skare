package game

// Client represents a two-way connection with the player's client.
type Client interface {
	// Stop stops the client running.
	Stop()
	// GetLine returns the next non-blank line of text from the player. The
	// second return value is false when the connection is closed.
	GetLine() (string, bool)
	// PutLine prints a line to the client with appropriate line endings.
	PutLine(string, ...any)
	// PutRaw prints a string to the client.
	PutRaw(string, ...any)
}
