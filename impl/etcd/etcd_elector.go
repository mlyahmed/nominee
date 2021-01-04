package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
)

// Elector ...
type Elector struct {
	*Etcd
	managedNode node.Node
	leaderEtcd  clientv3.GetResponse
	leaderSpec  *node.Spec
	election    Election
}

// NewElector ...
func NewElector(cl ConfigLoader) *Elector {
	cl.Load(context.Background())
	spec := cl.GetSpec()
	log = logger.G(context.Background()).WithFields(logrus.Fields{"elector": "etcd", "domain": spec.Domain, "cluster": spec.Cluster})
	racer := Elector{Etcd: NewEtcd(cl)}
	racer.failBackFn = func() error { return racer.connect(true) }
	return &racer
}

// Run ...
func (e *Elector) Run(n node.Node) error {
	log = log.WithFields(logrus.Fields{"daemon": n.GetDaemonName(), "node": n.GetName()})
	log.Infof("starting...")
	e.managedNode = n
	e.forwardStopChan()
	if err := e.connect(false); err != nil {
		return err
	}
	e.listenToTheConnectorSession()
	log.Infof("started.")
	return nil
}

func (e *Elector) UpdateLeader(leader *node.Spec) error {
	amICurrentlyTheLeader := e.amITheLeader()
	amITheNewLeader := leader.Name == e.managedNode.GetName()
	e.leaderSpec = leader
	if amITheNewLeader && amICurrentlyTheLeader {
		log.Infof("I stay the leader. Nothing to do.")
	} else if amITheNewLeader && !amICurrentlyTheLeader {
		log.Infof("promoting The Node...")
		if err := e.managedNode.Lead(e.Ctx, *e.leaderSpec); err != nil {
			e.Stonith()
		}
	} else if !amITheNewLeader && amICurrentlyTheLeader {
		_ = e.managedNode.Stonith(e.Ctx)
		e.Stonith()
	} else {
		if err := e.managedNode.Follow(e.Ctx, *e.leaderSpec); err != nil {
			_ = e.managedNode.Stonith(e.Ctx)
			e.Stonith()
		}
	}
	return nil
}

func (e *Elector) forwardStopChan() {
	go func() {
		<-e.managedNode.Stop()
		e.Stonith()
	}()
}

func (e *Elector) connect(reconnect bool) error {
	if reconnect {
		e.Reset()
	}

	if _, err := e.Connector.Connect(e.Ctx, e.ConfigSpec); err != nil {
		return err
	}

	if len(e.leaderEtcd.Kvs) > 0 {
		log.Infof("resume election...")
		e.election, _ = e.Connector.ResumeElection(e.Ctx, e.electionKey(), e.leaderEtcd)
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
		log.Infof("campaign as %v...", e.managedNode.GetName())
		if err := e.election.Campaign(e.Ctx, e.managedNode.GetSpec().Marshal()); err != nil {
			e.Stonith()
		}
	}()
}

func (e *Elector) observe() {
	go func() {
		o := e.election.Observe(e.Ctx)
		for leader := range o {
			e.leaderEtcd = leader
			spec := e.toNodeSpec(leader)
			_ = e.UpdateLeader(&spec)
		}
	}()
}

func (e *Elector) amITheLeader() bool {
	return e.leaderSpec != nil && e.leaderSpec.Name == e.managedNode.GetName()
}
