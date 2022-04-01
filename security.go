package main

type Security struct {
	AllowEnv  bool `yaml:"allow_env"`
	AllowArgs bool `yaml:"allow_args"`
}
