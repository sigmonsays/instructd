package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
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

type ExecResponse struct {
	Error    string
	ExitCode int
	Pid      int
	StdOut   string
	StdErr   string
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

	cd, err := me.findCommand(req.Id)
	if err != nil {
		me.sendError(w, r, "findCommand %s", err)
		return
	}

	log.Tracef("found command id:%v", cd.Id)

	ctx := context.Background()
	ret := &ExecResponse{}

	if len(cd.Cmd) == 0 && cd.Shell == "" {
		me.sendError(w, r, "shell or cmd required in request")
		return
	}

	if len(cd.Cmd) == 0 {
		cd.Cmd = []string{"sh", "-c", cd.Shell}
	}

	c := exec.CommandContext(ctx, cd.Cmd[0], cd.Cmd[1:]...)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()
	if err != nil {
		ret.Error = err.Error()
		// grab exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ret.ExitCode = exitError.ExitCode()
			ret.Pid = c.ProcessState.Pid()
		}
	}
	ret.StdOut = stdout.String()
	ret.StdErr = stderr.String()

	// send response
	rbuf, _ := json.Marshal(ret)
	w.Write(rbuf)

}

func (me *CommandHandler) findCommand(id string) (*CommandDetail, error) {
	for _, c := range me.Commands {
		if c.Id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("command not found")
}
