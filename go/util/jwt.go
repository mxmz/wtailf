package util

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type PubKeyJwtAuthorizer struct {
	pubKey *rsa.PublicKey
}

type JwtData struct {
	Sub string
	Iss string
}

func (a *PubKeyJwtAuthorizer) Authorize(r *http.Request) (*JwtData, error) {
	var t = strings.Split(r.Header.Get("Authorization"), " ")
	if len(t) != 2 {
		return nil, errors.New("400: Invalid Auhorization header")
	}
	var token = t[1]
	var data, err = a.decode(token)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (v *PubKeyJwtAuthorizer) decode(token string) (*JwtData, error) {
	decoded, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return v.pubKey, nil
	})
	if err != nil {
		return nil, err
	}
	var rv JwtData
	var claims jwt.MapClaims = decoded.Claims.(jwt.MapClaims)
	if sub, ok := claims["sub"].(string); ok && len(sub) > 0 {
		rv.Sub = sub
	} else {
		return nil, errors.New("400: Missing or invalid 'sub' claim")
	}
	if iss, ok := claims["iss"].(string); ok && len(iss) > 0 {
		rv.Iss = iss
	} else {
		return nil, errors.New("400: Missing or invalid 'iss' claim")
	}
	return &rv, nil
}

func NewPubKeyJwtAuthorizer(pubKeyPath string) (*PubKeyJwtAuthorizer, error) {
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, err
	}
	return &PubKeyJwtAuthorizer{pubKey: verifyKey}, nil
}
