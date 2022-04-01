package main

type CommandDetail struct {
	Id    string
	Shell string
	Cmd   []string
	Env   map[string]string
}

func (me *CommandDetail) Duplicate() *CommandDetail {
	cd := *me
	return &cd
}
