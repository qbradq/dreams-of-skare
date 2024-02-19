package game

import (
	"log"
	"time"
)

// Dream is a collection of interconnected scenes that represents the entire
// dream space the players explore.
type Dream struct {
	Creation      time.Time         // Creation time of the dream
	StartingScene string            // UUID of the starting scene
	Scenes        map[string]*Scene // Collection of all scenes in the dream
}

// NewDream creates a new Dream ready for use.
func NewDream() *Dream {
	d := &Dream{
		Creation: time.Now(),
		Scenes:   map[string]*Scene{},
	}
	s := NewScene()
	s.Name = "The Void"
	s.Description = "A featureless, formless void."
	d.StartingScene = s.UUID
	d.Scenes[s.UUID] = s
	return d
}

// InsertPlayer inserts the player's actor into the world by scene UUID.
func (d *Dream) InsertPlayer(a *Actor) bool {
	if a.Player == nil {
		log.Println("error: insert nil actor")
		return false
	}
	s, found := d.Scenes[a.Player.CurrentSceneUUID]
	if !found {
		log.Printf("missing scene uuid %s", a.Player.CurrentSceneUUID)
		return false
	}
	return s.AddActor(a)
}
