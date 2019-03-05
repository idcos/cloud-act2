//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package middleware

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-chi/chi/middleware"
	act2Log "idcos.io/cloud-act2/log"

	"log"
	"net/http"
	"os"
	"sync"

	"idcos.io/cloud-act2/config"
)

var (
	httpLogFile *os.File
	prevent     sync.Once
)

func CloseHttpLogFile() {
	if httpLogFile != nil {
		httpLogFile.Close()
	}
}

// not do current, use chi middleware
func Logger(next http.Handler) http.Handler {
	prevent.Do(func() {
		file, err := os.OpenFile(config.Conf.Logger.HttpLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("could not open http log file %s\n", config.Conf.Logger.HttpLogFile)
			return
		}
		httpLogFile = file
	})

	if httpLogFile != nil {
		RequestLogger := middleware.RequestLogger(
			&middleware.DefaultLogFormatter{
				Logger: log.New(httpLogFile, "", log.LstdFlags),
			},
		)

		return RequestLogger(next)
	} else {
		return next
	}
}

type Responser struct {
	w   http.ResponseWriter
	buf *bytes.Buffer
}

func (r *Responser) Header() http.Header {
	return r.w.Header()
}

func (r *Responser) Write(buf []byte) (int, error) {
	r.buf.Write(buf)
	return r.w.Write(buf)
}

func (r *Responser) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
}

func ResponseTracer(next http.Handler) http.Handler {

	logger := act2Log.L()

	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := &Responser{
			w:   w,
			buf: bytes.NewBufferString(""),
		}

		if logger.IsTrace() {
			t1 := time.Now()

			defer func() {
				t2 := time.Now()
				context := r.Context()
				requestKeyID := context.Value(middleware.RequestIDKey)

				logger.Trace("request info", "method", r.Method, "url", r.URL.String(), "host", r.Host, "headers", fmt.Sprintf("%#v", r.Header),
					"form", fmt.Sprintf("%#v", r.Form), "requestKeyId", requestKeyID, "response", ww.buf.String(),
					"time", fmt.Sprintf("%#v", t2.Sub(t1)))
			}()
		}

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
