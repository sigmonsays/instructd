package main

type CommandDetail struct {
	Id    string
	Shell string
	Cmd   []string
	Env   map[string]string
}
