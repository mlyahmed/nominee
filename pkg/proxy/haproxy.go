package proxy

import (
	"github.com/haproxytech/client-native/v2/configuration"
	"github.com/haproxytech/models/v2"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
)

type HAProxy struct {
	config *Config
	*configuration.Client
	currentTx *models.Transaction
	version   int64
	primary   nominee.Nominee
	standbies []nominee.Nominee
}

const (
	primaryBackend string = "be_primary"
	standbyBackend string = "be_standby"
)

func (proxy *HAProxy) Config() *Config {
	return proxy.config
}

func NewHAProxy(domain, cluster string) *HAProxy {
	proxy := HAProxy{
		config: &Config{
			Domain:  domain,
			Cluster: cluster,
		},
		Client: &configuration.Client{},
	}

	cgfFilePath := "/home/ahmed/data/projects/postgres-operator/labs/nominee/haproxy.cfg"
	if _, err := os.Stat(cgfFilePath); os.IsNotExist(err) {
		file, _ := os.Create(cgfFilePath)
		_ = file.Close()
	}
	confParams := configuration.ClientParams{
		ConfigurationFile:      cgfFilePath,
		Haproxy:                "/usr/sbin/haproxy",
		UseValidation:          true,
		PersistentTransactions: true,
		TransactionDir:         "/tmp/haproxy",
	}

	if err := proxy.Init(confParams); err != nil {
		panic(err)
	}

	version, err := proxy.GetVersion("")
	if err != nil {
		panic(err)
	}

	proxy.version = version
	proxy.startNewTx()
	proxy.removeAllServers()
	proxy.commitCurrentTx()

	return &proxy
}

func (proxy *HAProxy) PushNominees(nominees ...nominee.Nominee) error {
	proxy.startNewTx()
	for _, v := range nominees {
		proxy.removeServer(v.Name, primaryBackend)
		proxy.removeServer(v.Name, standbyBackend)
		proxy.createServer(standbyBackend, v)
	}
	proxy.standbies = append(proxy.standbies, nominees...)
	proxy.commitCurrentTx()
	return nil
}

func (proxy *HAProxy) PushLeader(leader nominee.Nominee) error {
	proxy.primary = leader
	proxy.startNewTx()
	proxy.removeServer(leader.Name, primaryBackend)
	proxy.removeServer(leader.Name, standbyBackend)
	proxy.createServer(primaryBackend, leader)
	proxy.commitCurrentTx()
	return nil
}

func (proxy *HAProxy) RemoveNominee(electionKey string) error {
	proxy.startNewTx()
	if proxy.primary.ElectionKey == electionKey {
		proxy.removePrimaryServer()
	} else {
		for k, v := range proxy.standbies {
			if v.ElectionKey == electionKey {
				proxy.standbies = append(proxy.standbies[:k], proxy.standbies[k+1:]...)
				proxy.removeServer(v.Name, standbyBackend)
				break
			}
		}

	}
	proxy.commitCurrentTx()
	return nil
}

func (proxy *HAProxy) removeAllServers() {
	proxy.removePrimaryServer()
	proxy.removeStandbyServers()
}

func (proxy *HAProxy) removePrimaryServer() {
	_, primary, err := proxy.GetServers(primaryBackend, proxy.currentTx.ID)
	if err != nil {
		panic(err)
	}
	for _, server := range primary {
		proxy.removeServer(server.Name, primaryBackend)
	}
}

func (proxy *HAProxy) removeStandbyServers() {
	_, standbies, err := proxy.GetServers(standbyBackend, proxy.currentTx.ID)
	if err != nil {
		panic(err)
	}
	for _, server := range standbies {
		proxy.removeServer(server.Name, standbyBackend)
	}
}

func (proxy *HAProxy) createServer(backend string, nominee nominee.Nominee) {
	weight := int64(100)
	if err := proxy.CreateServer(backend, &models.Server{
		Name:    nominee.Name,
		Address: nominee.Address,
		Port:    &nominee.Port,
		Check:   "enabled",
		Observe: "layer4",
		Weight:  &weight,
	}, proxy.currentTx.ID, 0); err != nil {
		panic(err)
	}
}

func (proxy *HAProxy) removeServer(name, backend string) {
	if _, _, err := proxy.GetServer(name, backend, proxy.currentTx.ID); err != nil {
		return
	}

	if err := proxy.DeleteServer(name, backend, proxy.currentTx.ID, 0); err != nil {
		panic(err)
	}
}

func (proxy *HAProxy) startNewTx() {
	tx, err := proxy.StartTransaction(proxy.version)
	if err != nil {
		panic(err)
	}
	proxy.currentTx = tx
}

func (proxy *HAProxy) commitCurrentTx() {
	_, err := proxy.CommitTransaction(proxy.currentTx.ID)
	if err != nil {
		panic(err)
	}
	proxy.version++
	proxy.currentTx = nil
}
