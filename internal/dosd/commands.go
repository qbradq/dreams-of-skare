package dosd

import (
	"fmt"
	"sort"
	"strings"
)

func executeCommand(l string, a *account) {
	parts := strings.SplitN(l, " ", 2)
	if len(parts) < 1 {
		return
	}
	d, found := cFnMap[parts[0]]
	if !found {
		a.Character.Player.Client.Put("Unknown command %s.")
		return
	}
	var cl string
	if len(parts) > 1 {
		cl = parts[1]
	}
	if a.Access < d.Al {
		a.Character.Player.Client.Put("You do not have access to command %s.")
		return
	}
	d.Fn(cl, a)
}

var cFnMap = map[string]*commandDescription{}

var cFns = []*commandDescription{}

type commandDescription struct {
	Name string                 // Primary command name
	Fn   func(string, *account) // Function to execute
	Al   accessLevel            // Minimum access level required
	Help string                 // Help text
}

func regCmd(ids []string, help string, al accessLevel, fn func(string, *account)) {
	if len(ids) < 1 {
		panic("no command names given")
	}
	d := &commandDescription{
		Name: ids[0],
		Fn:   fn,
		Al:   al,
		Help: help,
	}
	cFns = append(cFns, d)
	for _, id := range ids {
		if _, duplicate := cFnMap[id]; duplicate {
			panic(fmt.Sprintf("duplicate command function alias %s", id))
		}
		cFnMap[id] = d
	}
}

func init() {
	regCmd([]string{"/list"},
		"Lists all commands available to you.",
		alPlayer,
		func(s string, a *account) {
			for _, d := range cFns {
				if a.Access < d.Al {
					continue
				}
				a.Character.Player.Client.Put("% 17s - %s\n", d.Name, d.Help)
			}
		},
	)
	regCmd([]string{"/help"},
		"Displays help about the command.",
		alPlayer,
		func(s string, a *account) {
			d, found := cFnMap[s]
			if !found {
				a.Character.Player.Client.Put("Command %s not found.", s)
				return
			}
			if a.Access < d.Al {
				a.Character.Player.Client.Put("You do not have access to command %s.", d.Name)
				return
			}
			a.Character.Player.Client.Put("%s - %s\n", d.Name, d.Help)
		},
	)
	regCmd([]string{"/shutdown"},
		"Gracefully shuts down the server immediately.",
		alAdministrator,
		func(s string, a *account) {
			gracefulShutdown()
		},
	)
	sort.Slice(cFns, func(i, j int) bool {
		return cFns[i].Name < cFns[j].Name
	})
}
