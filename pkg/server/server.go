/*
Copyright 2024 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rkosegi/routeros2rest-bridge/pkg/types"
	"github.com/samber/lo"

	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
)

type Interface interface {
	io.Closer
	// Run runs infinitely
	Run() error
	// Init initializes server
	Init()
}

type rest struct {
	cfg    *types.Config
	server *http.Server
	logger *slog.Logger
	// pre-computed list of devices to send to API clients
	devices []*api.DeviceDetail
}

func (rs *rest) Close() error {
	return rs.server.Close()
}

func (rs *rest) specHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if data, err := api.PathToRawSpec(r.URL.Path)[r.URL.Path](); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		}
	}
}

func (rs *rest) Init() {
	rs.logger.Info("initializing server", "address", rs.cfg.Server.HTTPListenAddress)
	rs.devices = lo.Map(lo.Values(rs.cfg.Devices), func(dev *api.DeviceDetail, _ int) *api.DeviceDetail {
		return &api.DeviceDetail{
			Username: dev.Username,
			Password: "*********",
			Address:  dev.Address,
			Tls:      dev.Tls,
		}
	})
	r := mux.NewRouter()
	r.HandleFunc("/spec/opeanapi.v1.json", rs.specHandler())

	rs.server = &http.Server{
		Addr: rs.cfg.Server.HTTPListenAddress,
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
			}),
			handlers.AllowedOrigins(rs.cfg.Server.CorsConfig.AllowedOrigins),
			handlers.MaxAge(rs.cfg.Server.CorsConfig.MaxAge),
			handlers.AllowedHeaders([]string{"Content-Type"}),
		)(api.HandlerWithOptions(rs, api.GorillaServerOptions{
			BaseURL:    "/api/v1",
			BaseRouter: r,
			Middlewares: []api.MiddlewareFunc{
				loggingMiddleware(rs.logger.With("type", "access log")),
			},
		})),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func (rs *rest) Run() (err error) {
	defer func(rs *rest) {
		_ = rs.Close()
	}(rs)
	if rs.cfg.Server.HTTPTLSConfig != nil {
		return rs.server.ListenAndServeTLS(
			rs.cfg.Server.HTTPTLSConfig.TLSCertPath,
			rs.cfg.Server.HTTPTLSConfig.TLSKeyPath,
		)
	} else {
		return rs.server.ListenAndServe()
	}
}

type Opt func(*rest)

func WithLogger(logger *slog.Logger) Opt {
	return func(r *rest) {
		r.logger = logger
	}
}

func New(cfg *types.Config, opts ...Opt) Interface {
	r := &rest{cfg: cfg, logger: slog.Default()}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
