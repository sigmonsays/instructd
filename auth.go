package main

import (
	"errors"
	"net/http"
)

func NewAuth(creds map[string]*JwtCredential) *Auth {
	ret := &Auth{}

	if creds == nil || len(creds) == 0 {
		ret.DisableAuth = true
		return ret
	}

	backend := NewJwtAuth()
	ret.backend = backend
	for accessKey, cred := range creds {
		backend.Creds[accessKey] = cred
	}
	return ret
}

type Auth struct {
	backend     *JwtAuth
	DisableAuth bool
}

type JwtKey struct {
	AccessKey string
	SecretKey string
}

var ErrAccessKeyNotFound = errors.New("No such access key")

func (me *Auth) authenticate(r *http.Request) error {

	if me.DisableAuth {
		return nil
	}

	authRes, err := me.backend.ParseToken(r)
	if err != nil {
		return err
	}

	log.Tracef("auth results %+v", authRes)

	return nil
}
