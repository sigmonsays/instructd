package main

type CommandDetail struct {
	Id    string
	Shell string
	Cmd   []string
	Env   map[string]string
}

func (me *CommandDetail) Duplicate() *CommandDetail {
	cd := *me

	// copy env
	cd.Env = make(map[string]string, 0)
	for k, v := range me.Env {
		cd.Env[k] = v
	}

	log.Tracef("Duplicate return %#v", cd)
	return &cd
}
