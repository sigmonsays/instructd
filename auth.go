package main

import "errors"

type Auth struct {
	Keys map[string]*JwtKey
}

type JwtKey struct {
	AccessKey string
	SecretKey string
}

var ErrAccessKeyNotFound = errors.New("No such access key")

func (me *Auth) GetAccessKey(k string) (*JwtKey, error) {
	jwtKey, found := me.Keys[k]
	if found == false {
		return nil, ErrAccessKeyNotFound
	}
	return jwtKey, nil
}
