package util

import (
	"context"
	"net/http"
)

type AuthContext map[string]interface{}

type AuthFunc func(AuthContext, *http.Request) (AuthContext, error)

func ComposeAuthorizers(aa ...AuthFunc) AuthFunc {
	var f = func(c AuthContext, r *http.Request) (AuthContext, error) {
		var err error
		for _, a := range aa {
			c, err = a(c, r)
			if err != nil {
				return nil, err
			}
		}
		return c, nil
	}
	return f
}

type contextKey struct{}

var AuthContextKey = &contextKey{}

func AuthorizedHandlerBuilder(f AuthFunc) func(h http.HandlerFunc) http.HandlerFunc {
	return func(hndlr http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var c = AuthContext{}
			data, err := f(c, r)
			if err != nil {
				w.WriteHeader(403)
				w.Write([]byte(err.Error()))
				return
			}
			ctx := context.WithValue(r.Context(), AuthContextKey, data)
			hndlr(w, r.WithContext(ctx))
		}
	}
}

var NullAuthorizer AuthFunc = func(c AuthContext, r *http.Request) (AuthContext, error) {
	return c, nil
}
