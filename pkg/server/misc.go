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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/rkosegi/go-http-commons/output"
	"github.com/rkosegi/routeros2rest-bridge/pkg/api"
	"github.com/samber/lo"
	"gopkg.in/routeros.v2"
)

const (
	defaultPlainPort = "8728"
	defaultTLSPort   = "8729"
)

var (
	ErrCaAppend = errors.New("failed to append root CA certificate")
	out         = output.NewBuilder().Build()
)

type PathHandler func(dev *api.DeviceDetail, alias *api.AliasDetail, w http.ResponseWriter, r *http.Request)
type ItemHandler func(dev *api.DeviceDetail, alias *api.AliasDetail, id string, w http.ResponseWriter, r *http.Request)

func sendJson(w http.ResponseWriter, v interface{}) {
	out.SendWithStatus(w, v, http.StatusOK)
}

func (rs *rest) openConnection(dev *api.DeviceDetail) (net.Conn, error) {
	var (
		err     error
		host    string
		port    string
		rootCAs *x509.CertPool
	)
	timeout := time.Second * time.Duration(int64(*dev.Timeout))
	rs.logger.Debug("opening connection to device", "address", dev.Address, "timeout", timeout, "tls", dev.Tls)
	host, port, err = net.SplitHostPort(dev.Address)
	if err != nil {
		return nil, err
	}
	if dev.Tls == nil {
		if port == "" {
			port = defaultPlainPort
		}
		return net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	} else {
		if port == "" {
			port = defaultTLSPort
		}
		if dev.Tls != nil && dev.Tls.Ca != nil {
			var caCertPEM []byte
			rootCAs = x509.NewCertPool()
			if caCertPEM, err = os.ReadFile(*dev.Tls.Ca); err != nil {
				return nil, err
			}
			if ok := rootCAs.AppendCertsFromPEM(caCertPEM); !ok {
				return nil, ErrCaAppend
			}
		}
		return tls.DialWithDialer(&net.Dialer{
			Timeout: timeout,
		}, "tcp", net.JoinHostPort(host, port), &tls.Config{
			InsecureSkipVerify: dev.Tls.Verify,
			RootCAs:            rootCAs,
		})
	}
}

// withDevice creates client connection to device and pass it to consumer function
func (rs *rest) withDevice(dev *api.DeviceDetail, fn func(*routeros.Client) error) error {
	var (
		err  error
		conn net.Conn
		cl   *routeros.Client
	)
	conn, err = rs.openConnection(dev)
	if err != nil {
		return err
	}
	rs.logger.Debug("opened connection to device", "remote", conn.RemoteAddr(), "local", conn.LocalAddr())
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	if cl, err = routeros.NewClient(conn); err != nil {
		return err
	}
	if err = cl.Login(dev.Username, dev.Password); err != nil {
		return err
	}
	defer func() {
		rs.logger.Debug("closing client connection",
			"remote", conn.RemoteAddr().String(), "local", conn.LocalAddr().String())
		cl.Close()
	}()
	return fn(cl)
}

// consumeBodyAsCmds reads request body as JSON object and converts it to sequence of sentences,
// while prepending it with other set of sentences
func consumeBodyAsCmds(preCmds []string, r *http.Request) ([]string, error) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	} else {
		return append(preCmds, lo.Map(lo.Entries(body), func(entry lo.Entry[string, string], _ int) string {
			return fmt.Sprintf("=%s=%s", entry.Key, entry.Value)
		})...), nil
	}
}

func getItemCommands(path, id, action string) []string {
	return []string{
		fmt.Sprintf("%s/%s", path, action), fmt.Sprintf("?.id=%s", id),
	}
}
