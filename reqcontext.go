package main

import (
	"bytes"
	"fmt"
	"time"
)

type ReqContext struct {
	logbuf  bytes.Buffer
	started time.Time
	stopped time.Time
	dur     time.Duration
}

func NewReqContext() *ReqContext {
	me := &ReqContext{}
	return me
}

func (me *ReqContext) Append(value string) {
	fmt.Fprint(&me.logbuf, value)
}
func (me *ReqContext) Start() {
	me.started = time.Now()
}

func (me *ReqContext) Stop() {
	me.stopped = time.Now()
	me.dur = me.stopped.Sub(me.started)
}

func (me *ReqContext) DurationMs() int64 {
	return int64(me.dur.Milliseconds())
}

func (me *ReqContext) AppendKV(key string, value string) {
	fmt.Fprintf(&me.logbuf, " %s=%s", key, value)
}

func (me *ReqContext) Printf(f string, args ...interface{}) {
	fmt.Fprintf(&me.logbuf, " "+f, args...)
}

func (me *ReqContext) String() string {
	return me.logbuf.String()
}
