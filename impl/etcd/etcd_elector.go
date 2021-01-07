package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/election"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
)

// Elector ...
type Elector struct {
	*Etcd
	*election.DefaultElector
	leader   clientv3.GetResponse
	election Election
}

// NewElector ...
func NewElector(cl ConfigLoader) *Elector {
	cl.Load(context.Background())
	spec := cl.GetSpec()
	log = logger.G(context.Background()).WithFields(logrus.Fields{"elector": "etcd", "domain": spec.Domain, "cluster": spec.Cluster})
	elector := Elector{Etcd: NewEtcd(cl)}
	elector.failBackFn = func() error { return elector.connect(true) }
	return &elector
}

// Run ...
func (e *Elector) Run(n node.Node) error {
	log = log.WithFields(logrus.Fields{"daemon": n.GetDaemonName(), "node": n.GetName()})
	log.Infof("starting...")
	e.DefaultElector = election.NewElector(n)

	if err := e.connect(false); err != nil {
		return err
	}
	e.listenToTheConnectorSession()
	log.Infof("started.")
	return nil
}

func (e *Elector) connect(reconnect bool) error {
	if reconnect {
		e.Reset()
	}

	if _, err := e.Connector.Connect(e.Ctx, e.ConfigSpec); err != nil {
		return err
	}

	if len(e.leader.Kvs) > 0 {
		log.Infof("resume election...")
		e.election, _ = e.Connector.ResumeElection(e.Ctx, e.electionKey(), e.leader)
	} else {
		log.Infof("new election...")
		e.election, _ = e.Connector.NewElection(e.Ctx, e.electionKey())
	}

	e.campaign()
	e.observe()
	log.Infof("session created.")
	return nil
}

func (e *Elector) campaign() {
	go func() {
		log.Infof("campaign as %v...", e.Managed.GetName())
		if err := e.election.Campaign(e.Ctx, e.Managed.GetSpec().Marshal()); err != nil {
			e.Stonith(context.TODO())
		}
	}()
}

func (e *Elector) observe() {
	go func() {
		o := e.election.Observe(e.Ctx)
		for leader := range o {
			e.leader = leader
			spec := e.toNodeSpec(leader)
			_ = e.UpdateLeader(&spec)
		}
	}()
}
