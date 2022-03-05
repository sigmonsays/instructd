package main

import "net/http"

var defaultConf = `
http_addr: 127.0.0.1:8944
commands: []
# EOF
`

func main() {
	cfg := &Config{}
	cfg.LoadYamlBuffer([]byte(defaultConf))

	err := http.ListenAndServe(cfg.HTTPAddr, nil)
	if err != nil {
		ExitError("ListenAndServe %s: %s", cfg.HTTPAddr, err)
	}

}
