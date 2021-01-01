package haproxy

import (
	"context"
	"fmt"
	"github.com/haproxytech/client-native/v2/configuration"
	"github.com/haproxytech/models/v2"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
	"os/exec"
	"sync"
)

// HAProxy ...
type HAProxy struct {
	*configuration.Client
	currentTx *models.Transaction
	version   int64
	primary   nominee.NodeSpec
	standbies []nominee.NodeSpec
	mutex     *sync.Mutex
	ctx       context.Context
	cancel    func()
}

const (
	primaryBackend string = "be_primary"
	standbyBackend string = "be_standby"
)

// NewHAProxy ...
func NewHAProxy(config *HAProxyConfig) *HAProxy {
	proxy := HAProxy{Client: &configuration.Client{}, mutex: &sync.Mutex{}}
	proxy.ctx, proxy.cancel = context.WithCancel(context.Background())

	confParams := configuration.ClientParams{
		ConfigurationFile:      config.ConfigFile,
		Haproxy:                config.ExecFile,
		TransactionDir:         config.TxDir,
		UseValidation:          true,
		PersistentTransactions: true,
	}

	if err := proxy.Init(confParams); err != nil {
		panic(err)
	}

	version, err := proxy.GetVersion("")
	if err != nil {
		panic(err)
	}

	proxy.version = version

	proxy.mutex.Lock()
	defer proxy.mutex.Unlock()
	proxy.startTx()
	proxy.removeAllServers()
	proxy.commitTx()

	proxy.start(false)
	return &proxy
}

func (proxy *HAProxy) start(reload bool) {
	go func(reload bool) {
		if reload {
			proxy.cancel()
			proxy.ctx, proxy.cancel = context.WithCancel(context.Background())
		}
		command := exec.CommandContext(proxy.ctx, "/docker-entrypoint.sh", proxy.Haproxy, "-f", proxy.ConfigurationFile)
		command.Stdout, command.Stderr = os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Printf("Run of /docker-entrypoint.sh -> %v\n", err)
		}
	}(reload)
}

// PushNominees ...
func (proxy *HAProxy) PushNodes(nominees ...nominee.NodeSpec) error {
	proxy.mutex.Lock()
	defer proxy.mutex.Unlock()
	proxy.startTx()
	for _, v := range nominees {
		proxy.removeServer(v.Name, primaryBackend)
		proxy.removeServer(v.Name, standbyBackend)
		proxy.addServer(standbyBackend, v)
	}
	proxy.standbies = append(proxy.standbies, nominees...)
	proxy.commitTx()
	proxy.start(true)
	return nil
}

// PushLeader ...
func (proxy *HAProxy) PushLeader(leader nominee.NodeSpec) error {
	proxy.mutex.Lock()
	defer proxy.mutex.Unlock()
	proxy.primary = leader
	proxy.startTx()
	proxy.removeServer(leader.Name, primaryBackend)
	proxy.removeServer(leader.Name, standbyBackend)
	proxy.addServer(primaryBackend, leader)
	proxy.commitTx()
	proxy.start(true)
	return nil
}

// RemoveNominee ...
func (proxy *HAProxy) RemoveNode(electionKey string) error {
	proxy.mutex.Lock()
	defer proxy.mutex.Unlock()
	proxy.startTx()
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
	proxy.commitTx()
	proxy.start(true)
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

func (proxy *HAProxy) addServer(backend string, nominee nominee.NodeSpec) {
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

	fmt.Printf("server %s added to the backend %s \n", nominee.Name, backend)
}

func (proxy *HAProxy) removeServer(name, backend string) {
	if _, _, err := proxy.GetServer(name, backend, proxy.currentTx.ID); err != nil {
		return
	}

	if err := proxy.DeleteServer(name, backend, proxy.currentTx.ID, 0); err != nil {
		panic(err)
	}

	fmt.Printf("server %s removed from the backend %s \n", name, backend)
}

func (proxy *HAProxy) startTx() {
	tx, err := proxy.StartTransaction(proxy.version)
	if err != nil {
		panic(err)
	}
	proxy.currentTx = tx
}

func (proxy *HAProxy) commitTx() {
	_, err := proxy.CommitTransaction(proxy.currentTx.ID)
	if err != nil {
		panic(err)
	}
	proxy.version++
}
