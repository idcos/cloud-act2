//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package master

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/casbin/casbin"
	"idcos.io/cloud-act2/config"

	"idcos.io/cloud-act2/websocket"

	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"idcos.io/cloud-act2/server/common"
	"idcos.io/cloud-act2/server/master/handler"

	serverMiddleware "idcos.io/cloud-act2/server/middleware"
)

const (
	jsonContentType = "application/json"
	aclModel        = "etc/acl/model.conf"
	aclPolicy       = "etc/acl/policy.csv"
)

// NewRouter for routes
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(serverMiddleware.Heartbeat("/"))

	r.Use(serverMiddleware.Logger)
	r.Use(serverMiddleware.URLFormat)
	r.Use(serverMiddleware.ResponseTracer)

	// load the casbin model and policy from files, database is also supported.
	if config.Conf.Act2.ACL {
		aclModelFile := filepath.Join(config.Conf.ProjectPath, aclModel)
		aclPolicyFile := filepath.Join(config.Conf.ProjectPath, aclPolicy)
		fmt.Printf("acl model file %s acl policy file %s\n", aclModelFile, aclPolicyFile)
		e := casbin.NewEnforcer(aclModelFile, aclPolicyFile)
		r.Use(serverMiddleware.Authorizer(e))
	}

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(serverMiddleware.ApiLogBefore)
	//r.Use(middleware.Heartbeat("/ping"))

	// 注意：api部分有脱敏的日志处理，所以新增的api有敏感数据，并且需要记录日志，
	// 则需要在apilog的middleware中处理脱敏数据

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", common.Index)
		r.Get("/ping", common.Status)
		r.Get("/version", common.Version)
		r.Post("/register", handler.Register)
		r.Post("/master/heart", handler.MasterHeart)
		r.Post("/callback/test", handler.CallbackTest)
		r.Get("/system/heartbeat", handler.SystemHeartbeat)
		r.Post("/proxy/alloc", handler.ProxyAllocHandler)

		r.Route("/job", func(r chi.Router) {
			r.Use(serverMiddleware.AllowContentType(jsonContentType))
			r.Post("/result", handler.SaltResult)
			r.Post("/exec", handler.ExecByHostIPs)
			r.Post("/id/exec", handler.ExecByHostIDs)
			r.Post("/ip/exec", handler.ExecByHostIPs)
			r.Get("/record", handler.FindJobRecordByID)
			//r.Put("/exec/cmd/black", handler.AddCmdBlack)
			//r.Delete("/exec/cmd/black", handler.RemoveCmdBlack)
			r.Get("/record/page", handler.FindJobRecordByPage)
			r.Get("/record/result", handler.FindRecordResultByID)
			r.Get("/stat", handler.JobStat)
		})

		r.Route("/ws", func(r chi.Router) {
			r.Get("/record", websocket.JobStdout)
		})

		r.Route("/idc", func(r chi.Router) {
			r.Use(serverMiddleware.AllowContentType(jsonContentType))
			r.Get("/all", handler.FindIdcs)
			r.Get("/host", handler.IDCHosts)
			r.Delete("/proxy", handler.DelProxy)
			r.Get("/proxy", handler.FindIdcProxy)
			r.Get("/proxies", handler.FindAllProxy)
			r.Get("/host/all", handler.AllIDCHosts)
		})

		r.Route("/host", func(r chi.Router) {
			r.Use(serverMiddleware.AllowContentType(jsonContentType))
			r.Post("/ips", handler.HostListByIP)
			r.Post("/result/callback", handler.HostResult)
			r.Get("/result", handler.FindRecordResultsByID)
			r.Put("/entity", handler.UpdateEntityIDByHostID)
			r.Put("/proxy", handler.UpdateHostProxy)
			r.Get("/all/info", handler.FindAllHostInfo)
		})

		r.Route("/quick", func(r chi.Router) {
			r.Post("/install", handler.QuickInstallMinion)
		})

		r.Route("/act2", func(r chi.Router) {
			r.Put("/host", handler.UpdateEntityIDByHostID)
		})

		r.Route("/web/hook", func(r chi.Router) {
			r.Post("/", handler.AddWebHook)
			r.Put("/", handler.UpdateWebHook)
			r.Delete("/", handler.DeleteWebHook)
		})

		r.Route("/complex", func(r chi.Router) {
			r.Route("/file/migrate", func(r chi.Router) {
				r.Get("/download", handler.FileMigrateDownload)
				r.Post("/", handler.FileMigrate)
			})
		})
	})

	r.Route("/debug", func(r chi.Router) {

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.RequestURI+"/pprof/", 301)
		})
		r.HandleFunc("/vars", common.ExpVars)

		r.Route("/pprof", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, r.RequestURI+"/", 301)
			})

			r.HandleFunc("/", pprof.Index)
			r.HandleFunc("/symbol", pprof.Symbol)
			r.HandleFunc("/cmdline", pprof.Cmdline)
			r.HandleFunc("/profile", pprof.Profile)
			r.Handle("/block", pprof.Handler("block"))
			r.Handle("/heap", pprof.Handler("heap"))
			r.Handle("/goroutine", pprof.Handler("goroutine"))
			r.Handle("/threadcreate", pprof.Handler("threadcreate"))
		})
	})

	return r
}
