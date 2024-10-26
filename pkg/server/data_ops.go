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
	"fmt"
	"net/http"
	"strings"

	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
	"github.com/samber/lo"
	"gopkg.in/routeros.v2"
	"gopkg.in/routeros.v2/proto"
)

func (rs *rest) handlePath(writer http.ResponseWriter, request *http.Request, dev api.Device, alias api.Alias, handler PathHandler) {
	rs.logger.Debug("handlePath", "dev", dev, "alias", alias)
	var (
		d  *api.DeviceDetail
		a  *api.AliasDetail
		ok bool
	)
	if d, ok = rs.cfg.Devices[dev]; !ok {
		http.Error(writer, fmt.Sprintf("no such device: %v", dev), http.StatusNotFound)
		return
	}
	if a, ok = rs.cfg.Aliases[alias]; !ok {
		http.Error(writer, fmt.Sprintf("no such alias: %v", alias), http.StatusNotFound)
		return
	}
	handler(d, a, writer, request)
}

func (rs *rest) handleItem(writer http.ResponseWriter, request *http.Request, dev api.Device, alias api.Alias, id api.Id, handler ItemHandler) {
	rs.logger.Debug("handleItem", "dev", dev, "alias", alias, "id", id)
	rs.handlePath(writer, request, dev, alias, func(d *api.DeviceDetail, a *api.AliasDetail, w http.ResponseWriter, r *http.Request) {
		handler(d, a, id, writer, r)
	})
}

func (rs *rest) doGetById(cl *routeros.Client, path, id string, w http.ResponseWriter, r *http.Request, validResponse int) {
	if err := rs.withClient(cl, getItemCommands(path, id, "print"), func(re *routeros.Reply) {
		if len(re.Re) == 0 {
			http.NotFound(w, r)
		} else {
			sendJsonWithStatus(w, re.Re[0].Map, validResponse)
		}
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (rs *rest) withClient(cl *routeros.Client, cmds []string, fn func(re *routeros.Reply)) error {
	rs.logger.Debug("sending command to device", "sentences", strings.Join(cmds, ","))
	if re, err := cl.Run(cmds...); err != nil {
		rs.logger.Error("got error from device", "error", err)
		return err
	} else {
		rs.logger.Debug("got response from device", "re", re)
		fn(re)
	}
	return nil
}

func (rs *rest) listItemsHandler() PathHandler {
	return func(dev *api.DeviceDetail, alias *api.AliasDetail, w http.ResponseWriter, r *http.Request) {
		if err := rs.withDevice(dev, func(cl *routeros.Client) error {
			return rs.withClient(cl, []string{fmt.Sprintf("%s/print", alias.Path)}, func(re *routeros.Reply) {
				sendJson(w, lo.Map(re.Re, func(item *proto.Sentence, _ int) map[string]string {
					return item.Map
				}))
			})
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (rs *rest) getItemHandler() ItemHandler {
	return func(dev *api.DeviceDetail, alias *api.AliasDetail, id string, w http.ResponseWriter, r *http.Request) {
		if err := rs.withDevice(dev, func(cl *routeros.Client) error {
			rs.doGetById(cl, alias.Path, id, w, r, http.StatusOK)
			return nil
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (rs *rest) createHandler() PathHandler {
	return func(dev *api.DeviceDetail, alias *api.AliasDetail, w http.ResponseWriter, r *http.Request) {
		if !*alias.Create {
			http.NotFound(w, r)
			return
		}
		var (
			err  error
			cmds []string
		)
		if cmds, err = consumeBodyAsCmds([]string{fmt.Sprintf("%s/add", alias.Path)}, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = rs.withDevice(dev, func(cl *routeros.Client) error {
			return rs.withClient(cl, cmds, func(re *routeros.Reply) {
				rs.doGetById(cl, alias.Path, re.Done.List[0].Value, w, r, http.StatusCreated)
			})
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (rs *rest) deleteItemHandler() ItemHandler {
	return func(dev *api.DeviceDetail, alias *api.AliasDetail, id string, w http.ResponseWriter, r *http.Request) {
		if !*alias.Delete {
			http.NotFound(w, r)
			return
		}
		if err := rs.withDevice(dev, func(cl *routeros.Client) error {
			return rs.withClient(cl, getItemCommands(alias.Path, id, "remove"), func(re *routeros.Reply) {
				if re.Done.Word == "!done" {
					w.WriteHeader(http.StatusNoContent)
				} else {
					http.Error(w, "invalid response from device", http.StatusInternalServerError)
				}
			})
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (rs *rest) patchItemHandler() ItemHandler {
	return func(dev *api.DeviceDetail, alias *api.AliasDetail, id string, w http.ResponseWriter, r *http.Request) {
		if !*alias.Update {
			http.NotFound(w, r)
			return
		}
		var (
			err  error
			cmds []string
		)
		if cmds, err = consumeBodyAsCmds(getItemCommands(alias.Path, id, "set"), r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = rs.withDevice(dev, func(cl *routeros.Client) error {
			return rs.withClient(cl, cmds, func(re *routeros.Reply) {
				rs.doGetById(cl, alias.Path, re.Done.List[0].Value, w, r, http.StatusAccepted)
			})
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
