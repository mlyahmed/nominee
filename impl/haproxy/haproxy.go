package haproxy

import (
	"context"
	"fmt"
	"github.com/haproxytech/client-native/v2/configuration"
	"github.com/haproxytech/models/v2"
	"github/mlyahmed.io/nominee/pkg/base"
	"github/mlyahmed.io/nominee/pkg/node"
	"os"
	"os/exec"
	"sync"
)

// HAProxy ...
type HAProxy struct {
	*configuration.Client
	currentTx *models.Transaction
	version   int64

	mutex  *sync.Mutex
	ctx    context.Context
	cancel func()
}

func (proxy *HAProxy) Publish(leader *node.Spec, followers ...*node.Spec) error {
	proxy.mutex.Lock()
	defer proxy.mutex.Unlock()
	proxy.startTx()
	proxy.removeAllServers()

	if leader != nil {
		proxy.addServer(primaryBackend, leader)
	}

	for _, follower := range followers {
		proxy.addServer(standbyBackend, follower)
	}

	proxy.commitTx()
	proxy.start(true)
	return nil
}

const (
	primaryBackend string = "be_primary"
	standbyBackend string = "be_standby"
)

// NewHAProxy ...
func NewHAProxy(cl ConfigLoader) *HAProxy {
	cl.Load(context.Background())
	config := cl.GetSpec()
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

func (proxy *HAProxy) Done() base.DoneChan {
	return make(chan struct{})
}

func (proxy *HAProxy) Stonith(context.Context) {
	panic("implement me")
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

func (proxy *HAProxy) addServer(backend string, nod *node.Spec) {
	if nod == nil {
		return
	}

	weight := int64(100)
	if err := proxy.CreateServer(backend, &models.Server{
		Name:    nod.Name,
		Address: nod.Address,
		Port:    &nod.Port,
		Check:   "enabled",
		Observe: "layer4",
		Weight:  &weight,
	}, proxy.currentTx.ID, 0); err != nil {
		panic(err)
	}

	fmt.Printf("server %s added to the backend %s \n", nod.Name, backend)
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
