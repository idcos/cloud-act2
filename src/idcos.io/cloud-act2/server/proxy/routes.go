//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package proxy

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"idcos.io/cloud-act2/server/common"
	serverMiddleware "idcos.io/cloud-act2/server/middleware"
	"idcos.io/cloud-act2/server/proxy/handler"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(serverMiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(serverMiddleware.Heartbeat("/"))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", common.Index)
		r.Get("/ping", common.Status)
		r.Post("/execute", handler.Execute)
		r.Post("/sync/execute", handler.SyncExecute)
		r.Get("/file/reversed", handler.FileReversed)
		r.Put("/system/heartbeat", handler.Heartbeat)
		r.Route("/job", func(r chi.Router) {
			r.Post("/realtime", handler.RealTime)
			r.Get("/stat", handler.Stat)
			r.Post("/sync/callback", handler.SyncCallback)
		})
		r.Post("/salt/event", handler.HandleSaltEvent)
		r.Post("/host/ping", handler.HostPing)
		r.Route("/complex", func(r chi.Router) {
			r.Route("/file/migrate", func(r chi.Router) {
				r.Get("/pull", handler.GetRemoteFile)
				r.Get("/download", handler.FileMigrateDownload)
				r.Post("/notify", handler.NotifyDownload)
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
			r.HandleFunc("/cmdline", pprof.Cmdline)
			r.HandleFunc("/profile", pprof.Profile)
			r.HandleFunc("/symbol", pprof.Symbol)
			r.Handle("/block", pprof.Handler("block"))
			r.Handle("/heap", pprof.Handler("heap"))
			r.Handle("/goroutine", pprof.Handler("goroutine"))
			r.Handle("/threadcreate", pprof.Handler("threadcreate"))
		})
	})

	return r
}
