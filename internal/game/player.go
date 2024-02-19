package game

// Player contains the data and functionality specific to the player.
type Player struct {
	Client           Client `json:"-"` // Connected client
	CurrentSceneUUID string // UUID of the scene the player is in
}
