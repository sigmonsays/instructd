package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CommandHandler struct {
	Commands []*CommandDetail
}

func (me *CommandHandler) sendError(w http.ResponseWriter, r *http.Request, msg string, args ...interface{}) {
	fmt.Fprintf(w, "ERROR: "+msg, args...)
}

type ExecRequest struct {
	Id string
}

func (me *CommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		me.sendError(w, r, "Method must be POST")
		return
	}
	var req *ExecRequest

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		me.sendError(w, r, "ReadAll %s", err)
		return
	}

	err = json.Unmarshal(buf, &req)
	if err != nil {
		me.sendError(w, r, "ReadAll %s", err)
		return
	}

	log.Infof("Request %+v", req)

}
