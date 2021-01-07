package haproxy

import (
	"context"
	"github.com/haproxytech/client-native/v2/configuration"
	"github.com/haproxytech/models/v2"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"os"
	"os/exec"
	"sync"
)

// HAProxy ...
type HAProxy struct {
	*proxy.BasicProxy
	*configuration.Client
	currentTx *models.Transaction
	version   int64
	mutex     *sync.Mutex
	wg        *sync.WaitGroup
	status    proxy.Status
}

func (p *HAProxy) Publish(leader *node.Spec, followers ...*node.Spec) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.startTx()
	p.removeAllServers()

	if leader != nil {
		p.addServer(primaryBackend, leader)
	}

	for _, follower := range followers {
		p.addServer(standbyBackend, follower)
	}

	p.commitTx()
	p.start()
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
	haProxy := HAProxy{
		BasicProxy: proxy.NewBasicProxy(),
		Client:     &configuration.Client{},
		mutex:      &sync.Mutex{},
		wg:         &sync.WaitGroup{},
		status:     proxy.Stopped,
	}

	confParams := configuration.ClientParams{
		ConfigurationFile:      config.ConfigFile,
		Haproxy:                config.ExecFile,
		TransactionDir:         config.TxDir,
		UseValidation:          true,
		PersistentTransactions: true,
	}

	if err := haProxy.Init(confParams); err != nil {
		panic(err)
	}

	version, err := haProxy.GetVersion("")
	if err != nil {
		panic(err)
	}

	haProxy.version = version
	_ = haProxy.Publish(nil)

	go func() {
		haProxy.wg.Wait()
		haProxy.Stonith(haProxy.Ctx)
	}()

	return &haProxy
}

func (p *HAProxy) start() {
	if p.status == proxy.Started {
		p.Reset()
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		log := logger.G(p.Ctx)
		cmd := exec.CommandContext(p.Ctx, "/docker-entrypoint.sh", p.Haproxy, "-f", p.ConfigurationFile)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Start(); err != nil {
			log.Errorf("Failed to start %v", err)
			return
		}
		if p.status == proxy.Started {
			log.Infof("Restarted with pid %d", cmd.Process.Pid)
		} else {
			log.Infof("Started with pid %d", cmd.Process.Pid)
		}

		p.status = proxy.Started
		if err := cmd.Wait(); err != nil {
			log.Debugf("Stopped with pid %d because %v", cmd.Process.Pid, err)
		}
	}()
}

func (p *HAProxy) removeAllServers() {
	p.removePrimaryServer()
	p.removeStandbyServers()
}

func (p *HAProxy) removePrimaryServer() {
	_, primary, err := p.GetServers(primaryBackend, p.currentTx.ID)
	if err != nil {
		panic(err)
	}
	for _, server := range primary {
		p.removeServer(server.Name, primaryBackend)
	}
}

func (p *HAProxy) removeStandbyServers() {
	_, standbies, err := p.GetServers(standbyBackend, p.currentTx.ID)
	if err != nil {
		panic(err)
	}
	for _, server := range standbies {
		p.removeServer(server.Name, standbyBackend)
	}
}

func (p *HAProxy) addServer(backend string, nod *node.Spec) {
	if nod == nil {
		return
	}

	weight := int64(100)
	initAddr := "last,libc,none" // So it never fails on restart https://cbonte.github.io/haproxy-dconv/1.9/configuration.html#5.2-init-addr
	if err := p.CreateServer(backend, &models.Server{
		Name:     nod.Name,
		Address:  nod.Address,
		Port:     &nod.Port,
		Check:    "enabled",
		Observe:  "layer4",
		Weight:   &weight,
		InitAddr: &initAddr,
	}, p.currentTx.ID, 0); err != nil {
		panic(err)
	}

	logger.G(p.Ctx).Infof("server %s added to the backend %s", nod.Name, backend)
}

func (p *HAProxy) removeServer(name, backend string) {
	if _, _, err := p.GetServer(name, backend, p.currentTx.ID); err != nil {
		return
	}

	if err := p.DeleteServer(name, backend, p.currentTx.ID, 0); err != nil {
		panic(err)
	}

	logger.G(p.Ctx).Infof("server %s removed from the backend %s", name, backend)
}

func (p *HAProxy) startTx() {
	tx, err := p.StartTransaction(p.version)
	if err != nil {
		panic(err)
	}
	p.currentTx = tx
}

func (p *HAProxy) commitTx() {
	_, err := p.CommitTransaction(p.currentTx.ID)
	if err != nil {
		panic(err)
	}
	p.version++
}
