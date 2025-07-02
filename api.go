package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alessio/shellescape"
)

type CommandHandler struct {
	Auth     *Auth
	Security *Security
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

func (me *CommandHandler) readRequest(w http.ResponseWriter, r *http.Request) (*ExecRequest, error) {
	if r.Method == "POST" {
		return me.readPostRequest(w, r)
	}

	if r.Method == "GET" {
		return me.readGetRequest(w, r)
	}

	return nil, fmt.Errorf("Unsupported metthod: %s", r.Method)
}

func (me *CommandHandler) readPostRequest(w http.ResponseWriter, r *http.Request) (*ExecRequest, error) {
	req := &ExecRequest{}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	reqvals := NewMapVals(buf)

	err = json.Unmarshal(buf, &req)
	if err != nil {
		return nil, err
	}

	if req != nil {
		req.Values = reqvals
	}

	log.Infof("POST request %+v", req)
	return req, nil
}

func (me *CommandHandler) readGetRequest(w http.ResponseWriter, r *http.Request) (*ExecRequest, error) {
	req := &ExecRequest{
		Env: make(map[string]string, 0),
	}

	q := r.URL.Query()
	reqvals := NewMapValsFromUrlValues(q)

	// load exec request from url params
	req.Id = q.Get("id")
	req.Pwd = q.Get("pwd")
	stdout, err := strconv.ParseBool(q.Get("stdout"))
	if err == nil {
		req.StdOut = stdout
	}
	stderr, err := strconv.ParseBool(q.Get("stderr"))
	if err == nil {
		req.StdErr = stderr
	}
	for k, vals := range q {
		if strings.HasPrefix(k, "env.") == false {
			continue
		}
		if len(k) <= 4 {
			continue
		}
		// k is "env.FOO"
		k := k[4:]
		req.Env[k] = vals[0]
	}
	req.Args = q["args"]

	if req != nil {
		req.Values = reqvals
	}

	log.Infof("GET Request %+v", req)
	return req, nil
}

func (me *CommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	reqctx := NewReqContext()
	me.logStart(w, r, reqctx)

	defer func() {
		reqctx.Stop()
		reqctx.Printf("dur_ms=%d", reqctx.DurationMs())
		buf := reqctx.String()
		fmt.Printf("%s\n", buf)
	}()

	err := me.handleAuth(w, r)
	if err != nil {
		me.sendError(w, r, "Auth %s", err)
		return
	}

	req, err := me.readRequest(w, r)
	if err != nil {
		me.sendError(w, r, "readRequest %s", err)
		return
	}

	ocd, err := me.findCommand(req.Id)
	if err != nil {
		me.sendError(w, r, "findCommand %s", err)
		return
	}

	cd := ocd.Duplicate()

	log.Tracef("found command id:%v", cd.Id)

	reqctx.AppendKV("cmd", cd.Id)

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
		cd.Cmd = []string{"/usr/bin/env", "sh", "-c", cd.Shell}
	}
	cd.Cmd = append(cd.Cmd, req.Args...)

	log.Tracef("built command %+v", cd.Cmd)

	started := time.Now()

	c := exec.CommandContext(ctx, cd.Cmd[0], cd.Cmd[1:]...)

	for k, v := range req.Env {
		e := k + "=" + shellescape.Quote(v)
		log.Tracef("command env %s", e)
		c.Env = append(c.Env, e)
	}

	if req.Body != "" {
		bin := bytes.NewBufferString(req.Body)
		c.Stdin = bin
	}

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	if req.Values.HasValue("stdout") == false || req.StdOut {
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
	stopped := time.Now()
	cmdDur := stopped.Sub(started)

	ret.DurMs = int64(cmdDur.Milliseconds())
	ret.StdOut = stdout.String()
	ret.StdErr = stderr.String()

	// interpret response code
	httpCode := 0
	if ret.ExitCode == 0 {
		httpCode = 200
	} else {
		httpCode = 400
	}
	reqctx.Printf("http_code=%d", httpCode)
	w.WriteHeader(httpCode)

	// send response
	rbuf, _ := json.Marshal(ret)
	w.Write(rbuf)

}

func (me *CommandHandler) logStart(w http.ResponseWriter, r *http.Request, reqctx *ReqContext) {
	reqctx.Start()
	reqctx.Append("INSTRUCTD")
	reqctx.Printf("ts=%d", time.Now().Unix())
	reqctx.AppendKV("method", r.Method)
	reqctx.AppendKV("path", r.URL.Path)
	reqctx.AppendKV("client_ip", me.getClientIP(w, r))
}

func (me *CommandHandler) getClientIP(w http.ResponseWriter, r *http.Request) string {
	// prefer X-Forwarded-For
	fwdfor := r.Header.Get("X-Forwarded-For")
	if fwdfor != "" {
		return fwdfor
	}
	return r.RemoteAddr
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
