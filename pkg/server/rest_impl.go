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
	"net/http"

	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
)

func (rs *rest) ListAliases(w http.ResponseWriter, _ *http.Request) {
	sendJson(w, rs.cfg.Aliases)
}

func (rs *rest) ListDevices(w http.ResponseWriter, _ *http.Request) {
	sendJson(w, rs.devices)
}

func (rs *rest) ListItems(w http.ResponseWriter, r *http.Request, dev api.Device, alias api.Alias) {
	rs.handlePath(w, r, dev, alias, rs.listItemsHandler())
}

func (rs *rest) GetItem(w http.ResponseWriter, r *http.Request, dev api.Device, alias api.Alias, id api.Id) {
	rs.handleItem(w, r, dev, alias, id, rs.getItemHandler())
}

func (rs *rest) CreateItem(w http.ResponseWriter, r *http.Request, dev api.Device, alias api.Alias) {
	rs.handlePath(w, r, dev, alias, rs.createHandler())
}

func (rs *rest) DeleteItem(w http.ResponseWriter, r *http.Request, dev api.Device, alias api.Alias, id api.Id) {
	rs.handleItem(w, r, dev, alias, id, rs.deleteItemHandler())
}

func (rs *rest) PatchItem(w http.ResponseWriter, r *http.Request, dev api.Device, alias api.Alias, id api.Id) {
	rs.handleItem(w, r, dev, alias, id, rs.patchItemHandler())
}
