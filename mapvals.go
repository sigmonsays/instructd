package main

import "encoding/json"

func NewMapVals(buf []byte) *MapVals {
	ret := &MapVals{}
	mapvals := make(map[string]interface{}, 0)
	json.Unmarshal(buf, &mapvals)
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
