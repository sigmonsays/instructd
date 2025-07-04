package main

import (
	"flag"
	"net/http"

	gologging "github.com/sigmonsays/go-logging"
)

var defaultConf = `
http_addr: 127.0.0.1:8944
commands: []
# EOF
`

func main() {
	configfile := "/etc/instructd.yaml"
	loglevel := "INFO"
	flag.StringVar(&configfile, "config", configfile, "specify config file")
	flag.StringVar(&loglevel, "loglevel", loglevel, "specify log level")
	flag.Parse()

	gologging.SetLogLevel(loglevel)

	cfg := &Config{}
	cfg.LoadYamlBuffer([]byte(defaultConf))
	cfg.LoadYaml(configfile)

	cfg.PrintConfig()

	api := &CommandHandler{
		Commands: cfg.Commands,
		Auth:     NewAuth(cfg.Auth),
		Security: cfg.Security,
	}

	if cfg.Security.AllowEnv {
		log.Infof("Security notice, passing env via HTTP is allowed")
	}
	if cfg.Security.AllowArgs {
		log.Infof("Security notice, passing args via HTTP is allowed")
	}

	log.Infof("Listening at %s", cfg.HTTPAddr)
	err := http.ListenAndServe(cfg.HTTPAddr, api)
	if err != nil {
		ExitError("ListenAndServe %s: %s", cfg.HTTPAddr, err)
	}

}
