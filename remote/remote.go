// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// xyzzy100

// Package remote integrates the remote features of Viper.
package remote

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"strings"

	crypt "github.com/sagikazarmark/crypt/config"
	goetcdv3 "go.etcd.io/etcd/client/v3"

	"github.com/spf13/viper"
)

// goetcdv3 "go.etcd.io/etcd/client/v3"

type remoteConfigProvider struct{}

func (rc remoteConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (rc remoteConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	resp, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(resp), nil
}

func (rc remoteConfigProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, nil
	}
	quit := make(chan bool)
	quitwc := make(chan bool)
	viperResponsCh := make(chan *viper.RemoteResponse)
	cryptoResponseCh := cm.Watch(rp.Path(), quit)
	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func(cr <-chan *crypt.Response, vr chan<- *viper.RemoteResponse, quitwc <-chan bool, quit chan<- bool) {
		for {
			select {
			case <-quitwc:
				quit <- true
				return
			case resp := <-cr:
				vr <- &viper.RemoteResponse{
					Error: resp.Error,
					Value: resp.Value,
				}
			}
		}
	}(cryptoResponseCh, viperResponsCh, quitwc, quit)

	return viperResponsCh, quitwc
}

func getConfigManager(rp viper.RemoteProvider) (cm crypt.ConfigManager, err error) {
	endpoints := strings.Split(rp.Endpoint(), ";")
	if rp.SecretKeyring() != "" {
		var kr *os.File
		kr, err = os.Open(rp.SecretKeyring())
		if err != nil {
			return nil, err
		}
		defer kr.Close()
		switch rp.Provider() {
		case "etcd":
			cm, err = crypt.NewEtcdConfigManager(endpoints, kr)
		case "etcd3":
			var username, password string
			endpoints, username, password, err = parseEndpointUsers(endpoints)
			if err != nil {
				return nil, err
			} else if username == "" {
				cm, err = crypt.NewEtcdV3ConfigManager(endpoints, kr)
			} else {
				// xyzzy100
				// cm, err = crypt.NewStandardEtcdConfigManagerFromConfig(goetcdv3.Config{
				cm, err = crypt.NewStandardEtcdV3ConfigManagerFromConfig(goetcdv3.Config{
					Endpoints: endpoints,
					Username:  username,
					Password:  password,
				})
			}
		case "firestore":
			cm, err = crypt.NewFirestoreConfigManager(endpoints, kr)
		case "nats":
			cm, err = crypt.NewNatsConfigManager(endpoints, kr)
		default:
			cm, err = crypt.NewConsulConfigManager(endpoints, kr)
		}
	} else {
		switch rp.Provider() {
		case "etcd":
			cm, err = crypt.NewStandardEtcdConfigManager(endpoints)
		case "etcd3":
			var username, password string
			endpoints, username, password, err = parseEndpointUsers(endpoints)
			if err != nil {
				return nil, err
			} else if username == "" {
				cm, err = crypt.NewStandardEtcdV3ConfigManager(endpoints)
			} else {
				// xyzzy100
				cm, err = crypt.NewStandardEtcdV3ConfigManagerFromConfig(goetcdv3.Config{
					Endpoints: endpoints,
					Username:  username,
					Password:  password,
				})
			}
		case "firestore":
			cm, err = crypt.NewStandardFirestoreConfigManager(endpoints)
		case "nats":
			cm, err = crypt.NewStandardNatsConfigManager(endpoints)
		default:
			cm, err = crypt.NewStandardConsulConfigManager(endpoints)
		}
	}
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// endpoints, username, password := parseEndpointUsers(endpoints)
func parseEndpointUsers(endpoints []string) (returnEndpoints []string, username, password string, err error) {
	found := false
	for _, endpoint := range endpoints {
		// if it is of the form, "etcd://un:pw@host:port" - then
		if strings.HasPrefix(endpoint, "etcd://") {
			found = true
			uu, err := url.Parse(endpoint)
			if err != nil {
				return endpoints, "", "", err
			}
			if username == "" { // just use the 1st one found
				username = uu.User.Username()
				if username != "" {
					password, _ = uu.User.Password()
				}
			}
			returnEndpoints = append(returnEndpoints, uu.Host)
		} else {
			returnEndpoints = append(returnEndpoints, endpoint)
		}
	}
	if found {
		return
	}
	returnEndpoints = endpoints
	return
}

func init() {
	viper.RemoteConfig = &remoteConfigProvider{}
}
