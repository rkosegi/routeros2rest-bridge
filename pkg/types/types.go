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

package types

import (
	"errors"
	"fmt"

	"dario.cat/mergo"
	ccfg "github.com/rkosegi/go-http-commons/config"
	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
)

var (
	vFalse     = false
	defTimeout = float32(30)
	defDevice  = &api.DeviceDetail{
		Timeout: &defTimeout,
	}
	defAlias = &api.AliasDetail{
		Create: &vFalse,
		Update: &vFalse,
		Delete: &vFalse,
	}
	defServerConfig = ccfg.ServerConfig{
		ListenAddress: "0.0.0.0:22003",
		Cors: &ccfg.CorsConfig{
			// if you run this in default config, you most likely come from
			// different origin then http://localhost:22003 (or whatever address is this running on).
			// Be sure to set something sane to fit your deployment.
			AllowedOrigins: []string{"*"},
			MaxAge:         1200,
		},
	}
)

type Config struct {
	Server  ccfg.ServerConfig `yaml:"server"`
	Aliases map[string]*api.AliasDetail
	Devices map[string]*api.DeviceDetail
}

func (c *Config) Normalize() error {
	var err error

	if err = mergo.Merge(&c.Server, defServerConfig); err != nil {
		return err
	}
	if len(c.Aliases) == 0 {
		return errors.New("no aliases defined")
	}
	for name, alias := range c.Aliases {
		alias.Name = &name
		if len(alias.Path) == 0 {
			return fmt.Errorf("alias '%s' is missing path", name)
		}
		if err = mergo.Merge(alias, defAlias); err != nil {
			return err
		}
	}
	if len(c.Devices) == 0 {
		return errors.New("no device defined")
	}
	for name, device := range c.Devices {
		device.Name = &name
		if len(device.Username) == 0 {
			return fmt.Errorf("device '%s' is missing username", name)
		}
		if len(device.Password) == 0 {
			return fmt.Errorf("device '%s' is missing password", name)
		}
		if len(device.Address) == 0 {
			return fmt.Errorf("device '%s' is missing address", name)
		}
		if err = mergo.Merge(device, defDevice); err != nil {
			return err
		}
	}
	return c.Server.Check()
}
