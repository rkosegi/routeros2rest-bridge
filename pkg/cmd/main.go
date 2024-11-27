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

package main

import (
	"errors"
	"flag"
	"net/http"
	"os"

	"github.com/rkosegi/routeros2rest-bridge/pkg/server"
	"github.com/rkosegi/routeros2rest-bridge/pkg/types"
	xlog "github.com/rkosegi/slog-config"
	"gopkg.in/yaml.v3"
)

func main() {
	var (
		cfgFile string
		err     error
		cfg     *types.Config
	)
	sc := xlog.MustNew("info", xlog.LogFormatLogFmt)
	flag.StringVar(&cfgFile, "config", "config.yaml", "config file")
	flag.Var(&sc.Level, "log-level", "log level")
	flag.Var(&sc.Format, "log-format", "log format")
	flag.Parse()
	logger := sc.Logger()

	logger.Debug("loading config", "file", cfgFile)
	if cfg, err = loadConfig(cfgFile); err != nil {
		logger.Error("error loading config", "error", err)
		os.Exit(1)
	}
	srv := server.New(cfg, server.WithLogger(logger))
	srv.Init()
	if err = srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("error while running server", "error", err)
		os.Exit(2)
	}
}

func loadConfig(cfgFile string) (*types.Config, error) {
	var (
		cfg  types.Config
		err  error
		data []byte
	)
	if data, err = os.ReadFile(cfgFile); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err = cfg.Normalize(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
