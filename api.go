package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/alessio/shellescape"
)

type CommandHandler struct {
	Auth     *Auth
	Security *Security
	Commands []*CommandDetail
}

type ExecRequest struct {
	Id     string            `json:"id"`
	Body   string            `json:"body"`
	Pwd    string            `json:"pwd"`
	StdOut bool              `json:"stdout"`
	StdErr bool              `json:"stderr"`
	Env    map[string]string `json:"env"`
	Args   []string          `json:"args"`
}

type ExecResponse struct {
	Error    string `json:"error"`
	ExitCode int    `json:"exit_code"`
	Pid      int    `json:"pid"`
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
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

func (me *CommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		me.sendError(w, r, "Method must be POST")
		return
	}

	err := me.handleAuth(w, r)
	if err != nil {
		me.sendError(w, r, "Auth %s", err)
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

	if me.Security.AllowEnv == false && len(req.Env) > 0 {
		me.sendError(w, r, "passing env not allowed")
		return
	}
	if me.Security.AllowArgs == false && len(req.Args) > 0 {
		me.sendError(w, r, "passing args not allowed")
		return
	}

	if len(cd.Cmd) == 0 {
		cd.Cmd = []string{"sh", "-c", cd.Shell}
	}
	cd.Cmd = append(cd.Cmd, req.Args...)

	c := exec.CommandContext(ctx, cd.Cmd[0], cd.Cmd[1:]...)

	for k, v := range req.Env {
		c.Env = append(c.Env, k+"="+shellescape.Quote(v))
	}

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
	c.Dir = req.Pwd

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

func (me *CommandHandler) handleAuth(w http.ResponseWriter, r *http.Request) error {

	jt := ExtractToken(r)
	log.Tracef("Authenticating request with jwt-token:%q", jt)

	err := me.Auth.Authenticate(r)
	if err != nil {
		return err
	}

	return nil
}
