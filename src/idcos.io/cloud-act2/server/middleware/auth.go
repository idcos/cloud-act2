//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/casbin/casbin"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/server/common"
)

var (
	errUser        = errors.New("invalid user")
	errAuthInvalid = errors.New("invalid authorization data")
	errACLAuth     = errors.New("unknown ACL auth method")
)

var (
	excludePaths = []string{"/api/v1/host/result/callback", "/api/v1/callback/test"}
)

func findACLUser(user string) *config.ACLUser {
	for _, ACLUser := range config.Conf.ACLUser {
		if user == ACLUser.Name {
			return &ACLUser
		}
	}
	return nil
}

func basicAuth(w http.ResponseWriter, r *http.Request) (string, []string, error) {
	user, password, ok := r.BasicAuth()
	if !ok {
		return "", nil, errAuthInvalid
	}

	// 判定用户名和密码是否合规
	aclUser := findACLUser(user)
	if aclUser == nil {
		fmt.Printf("not found user %s\n", user)
		return "", nil, errUser
	}

	if aclUser.Password != password {
		fmt.Printf("invalid password %s, %s\n", password, aclUser.Password)
		return "", nil, errUser
	}

	return user, aclUser.Role, nil
}

func auth(w http.ResponseWriter, r *http.Request) (string, []string, error) {
	switch config.Conf.Act2.ACLAuth {
	case "basic":
		return basicAuth(w, r)
	default:
		// 未知的异常错误
		return "", nil, errACLAuth
	}
}

// Authorizer for casbin
func Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {

	// 任意一个角色通过，即为通过
	isValidUser := func(user string, roles []string, path string, method string) bool {
		for _, role := range roles {
			if e.Enforce(role, path, method) {
				return true
			}
		}
		return false
	}

	inExcludePath := func(path string, excludePaths []string) bool {
		for _, excludePath := range excludePaths {
			if path == excludePath {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		var fn func(w http.ResponseWriter, r *http.Request)
		if config.Conf.Act2.ACL {
			fn = func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				if inExcludePath(path, excludePaths) {
					next.ServeHTTP(w, r)
					return
				}

				user, roles, err := auth(w, r)
				if err != nil {
					common.HandleError(w, err)
					return
				}

				method := r.Method

				if isValidUser(user, roles, path, method) {
					ctx := r.Context()
					ctx = context.WithValue(ctx, string("user"), user)
					ctx = context.WithValue(ctx, string("roles"), roles)
					r = r.WithContext(ctx)

					next.ServeHTTP(w, r)
				} else {
					http.Error(w, http.StatusText(403), 403)
				}
			}
		} else {
			fn = func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}
		}

		return http.HandlerFunc(fn)
	}
}
