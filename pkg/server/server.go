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
	"github.com/rkosegi/go-http-commons/middlewares"
	"github.com/rkosegi/go-http-commons/openapi"
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

func (rs *rest) Init() {
	rs.logger.Info("initializing server", "address", rs.cfg.Server.ListenAddress)
	rs.devices = lo.Map(lo.Values(rs.cfg.Devices), func(dev *api.DeviceDetail, _ int) *api.DeviceDetail {
		return &api.DeviceDetail{
			Username: dev.Username,
			Password: "*********",
			Address:  dev.Address,
			Tls:      dev.Tls,
		}
	})
	r := mux.NewRouter()
	r.HandleFunc("/spec/opeanapi.v1.json", openapi.SpecHandler(api.PathToRawSpec))

	rs.server = &http.Server{
		Addr: rs.cfg.Server.ListenAddress,
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
			}),
			handlers.AllowedOrigins(rs.cfg.Server.Cors.AllowedOrigins),
			handlers.MaxAge(rs.cfg.Server.Cors.MaxAge),
			handlers.AllowedHeaders([]string{"Content-Type"}),
		)(api.HandlerWithOptions(rs, api.GorillaServerOptions{
			BaseURL:    "/api/v1",
			BaseRouter: r,
			Middlewares: []api.MiddlewareFunc{
				middlewares.NewLoggingBuilder().WithLogger(rs.logger).Build(),
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
	return rs.cfg.Server.RunForever(rs.server)
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
