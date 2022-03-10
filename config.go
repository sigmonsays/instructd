package main

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HTTPAddr string                    `yaml:"http_addr"`
	Verbose  bool                      `yaml:"verbose"`
	Commands []*CommandDetail          `yaml:"commands"`
	Auth     map[string]*JwtCredential `yaml:"auth"`
}

func (c *Config) LoadYaml(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(nil)
	_, err = b.ReadFrom(f)
	if err != nil {
		return err
	}

	if err := c.LoadYamlBuffer(b.Bytes()); err != nil {
		return err
	}

	if err := c.FixupConfig(); err != nil {
		return err
	}

	return nil
}

func (c *Config) LoadYamlBuffer(buf []byte) error {
	err := yaml.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) FixupConfig() error {

	return nil
}

func (c *Config) PrintConfig() error {
	buf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", buf)
	return nil
}
