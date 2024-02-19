package game

import "github.com/qbradq/dreams-of-skare/internal/util"

// Scene represents one scene or area of the dream.
type Scene struct {
	UUID        string   // Unique ID of the scene
	Name        string   // Descriptive name
	Description string   // Descriptive text
	Actors      []*Actor // List of all actors in the scene
	Items       []*Item  // List of all items in the scene
}

// NewScene creates a new Scene ready for use.
func NewScene() *Scene {
	return &Scene{
		UUID: util.NewUUID(),
	}
}

// AddActor adds the actor to the scene's list of actors.
func (s *Scene) AddActor(a *Actor) bool {
	s.Actors = append(s.Actors, a)
	return true
}

// RemoveActor removes the actor from the scene's list of actors.
func (s *Scene) RemoveActor(a *Actor) bool {
	idx := -1
	for i, o := range s.Actors {
		if o == a {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	copy(s.Actors[idx:], s.Actors[idx+1:])
	s.Actors[len(s.Actors)-1] = nil
	s.Actors = s.Actors[:len(s.Actors)-1]
	return true
}
