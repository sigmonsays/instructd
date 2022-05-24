package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func NewMapVals(buf []byte) *MapVals {
	ret := &MapVals{}
	mapvals := make(map[string]interface{}, 0)
	json.Unmarshal(buf, &mapvals)
	ret.m = mapvals
	return ret
}
func NewMapValsFromUrlValues(q url.Values) *MapVals {
	ret := &MapVals{}
	mapvals := make(map[string]interface{}, 0)
	for k, v := range q {
		mapvals[k] = v[0]
	}
	ret.m = mapvals
	return ret
}

type MapVals struct {
	m map[string]interface{}
}

func (me *MapVals) HasValue(name string) bool {
	_, found := me.m[name]
	return found
}

func (me *MapVals) String() string {
	return fmt.Sprintf("%+v", me.m)
}
