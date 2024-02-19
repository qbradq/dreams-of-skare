package dosd

import (
	"sync"
	"time"

	"github.com/qbradq/dreams-of-skare/internal/game"
	"github.com/qbradq/dreams-of-skare/internal/util"
)

var accounts = map[string]*account{}

var accountsLock sync.RWMutex

type accessLevel int

const (
	alPlayer        accessLevel = 0
	alModerator     accessLevel = 1
	alDreamMaster   accessLevel = 2
	alAdministrator accessLevel = 3
)

// account holds all of the user-related information.
type account struct {
	Access       accessLevel // Level of access
	Created      time.Time   // Time of account creation
	Username     string      // Username of the account
	PasswordHash string      // Hash of the user's password
	Character    *game.Actor // The player's character
}

func accountLogin(c client) *account {
	c.Put("\nAccount Name: ")
	name, ok := c.GetLine()
	if !ok {
		return nil
	}
	accountsLock.RLock()
	a, found := accounts[name]
	accountsLock.RUnlock()
	if !found {
		c.Put("Creating new account \"%s\".\n", name)
		c.Put("Password: ")
		pass1, ok := c.GetLine()
		if !ok {
			return nil
		}
		c.Put("Confirm Password: ")
		pass2, ok := c.GetLine()
		if !ok {
			return nil
		}
		if pass1 != pass2 {
			c.Put("Passwords did not match. Disconnecting.\n")
			return nil
		}
		a = &account{
			Created:      time.Now(),
			Username:     name,
			PasswordHash: util.Hash(pass1),
			Character: &game.Actor{
				Player: &game.Player{
					CurrentSceneUUID: dream.StartingScene,
				},
				Name:        name,
				Description: "Describe your character.",
			},
		}
		accountsLock.Lock()
		if len(accounts) < 1 {
			c.Put("This is the first account created on the server. Granting administration rights.\n")
			a.Access = alAdministrator
		}
		accounts[name] = a
		accountsLock.Unlock()
		c.Put("Created new account \"%s\". Welcome dreamer!\n", name)
	} else {
		c.Put("Password: ")
		pass, ok := c.GetLine()
		if !ok {
			return nil
		}
		if a.PasswordHash != util.Hash(pass) {
			c.Put("Bad password for account \"%s\". Disconnecting.\n", name)
			return nil
		}
		if a.Character.Player.Client != nil {
			c.Put("Someone else was already logged into this account. Disconnecting both.")
			c.Stop()
			a.Character.Player.Client.Stop()
			return nil
		}
		c.Put("Welcome dreamer!\n")
	}
	a.Character.Player.Client = c
	return a
}
