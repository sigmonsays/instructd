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

type ErrorResponse struct {
	Error string `json:"error"`
}

func (me *CommandHandler) sendError(w http.ResponseWriter, r *http.Request, msg string, args ...interface{}) {
	ret := &ErrorResponse{}
	ret.Error = fmt.Sprintf("ERROR: "+msg, args...)
	buf, _ := json.Marshal(ret)
	w.Write(buf)
	log.Tracef("sendError %s", buf)
}

type ExecRequest struct {
	Id     string `json:"id"`
	Body   string `json:"body"`
	StdOut bool   `json:"stdout"`
	StdErr bool   `json:"stderr"`
}

type ExecResponse struct {
	Error    string `json:"error"`
	ExitCode int    `json:"exit_code"`
	Pid      int    `json:"pid"`
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
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

	reqvals := NewMapVals(buf)

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

	if req.Body != "" {
		bin := bytes.NewBufferString(req.Body)
		c.Stdin = bin
	}

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	if reqvals.HasValue("stdout") == false || req.StdOut {
		c.Stdout = &stdout
	}
	if req.StdErr {
		c.Stderr = &stderr
	}

	for k, v := range cd.Env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%q", k, v))
	}

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
