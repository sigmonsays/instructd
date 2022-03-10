package main

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

func NewJwtAuth() *JwtAuth {
	ret := &JwtAuth{}
	ret.Creds = make(map[string]*JwtCredential, 0)
	return ret
}

type JwtAuth struct {
	Creds map[string]*JwtCredential
}

type JwtCredential struct {
	SecretKey string `yaml:"secret_key"`
	Disabled  bool   `yaml:"disabled"`
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

type JwtToken struct {
	AccessKey string
}

func (me *JwtAuth) ParseToken(r *http.Request) (*JwtToken, error) {
	tokenString := ExtractToken(r)
	ret := &JwtToken{}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// get access key
		interfaceValue, ok := token.Header["kid"]
		if ok == false {
			return nil, fmt.Errorf("kid not found")
		}
		accessKey := interfaceValue.(string)
		log.Tracef("access key %q", accessKey)
		ret.AccessKey = accessKey

		// fetch secret
		cred, found := me.Creds[accessKey]
		if found == false {
			return nil, fmt.Errorf("invalid access key %q", accessKey)
		}
		if cred.Disabled {
			return nil, fmt.Errorf("disabled access key %q", accessKey)
		}
		secret := []byte(cred.SecretKey)

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	log.Tracef("claims %+v", token.Claims)
	return ret, nil
}
