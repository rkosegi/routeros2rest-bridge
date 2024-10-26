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
	"testing"

	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestConfigNormalize(t *testing.T) {
	var dummy = "dummy"
	var (
		err error
		c   *Config
	)
	c = &Config{}
	err = c.Normalize()
	assert.Error(t, err)
	c.Aliases = map[string]*api.AliasDetail{
		"good": {
			Name: &dummy,
			Path: "/system/packages",
		},
	}
	c.Devices = map[string]*api.DeviceDetail{
		"dev1": {
			Name:     &dummy,
			Username: "admin",
			Password: "admin",
			Address:  "10.11.12.13",
		},
	}

	err = c.Normalize()
	assert.NoError(t, err)
}
