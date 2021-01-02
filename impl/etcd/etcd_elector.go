package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/logger"
	node2 "github/mlyahmed.io/nominee/pkg/node"
)

// Elector ...
type Elector struct {
	*Etcd
	node     node2.Node
	leader   clientv3.GetResponse
	election Election
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
func (racer *Elector) Run(node node2.Node) error {
	log = log.WithFields(logrus.Fields{"daemon": node.GetDaemonName(), "node": node.GetName()})
	log.Infof("starting...")
	racer.node = node
	racer.forwardStopChan()
	if err := racer.connect(false); err != nil {
		return err
	}
	racer.listenToTheConnectorSession()
	log.Infof("started.")
	return nil
}

func (racer *Elector) forwardStopChan() {
	go func() {
		racer.StopChan <- <-racer.node.Stop()
	}()
}

func (racer *Elector) connect(reconnect bool) error {
	if reconnect {
		racer.Reset()
	}

	if _, err := racer.Connector.Connect(racer.Ctx, racer.ConfigSpec); err != nil {
		return err
	}

	if len(racer.leader.Kvs) > 0 {
		log.Infof("resume election...")
		racer.election, _ = racer.Connector.ResumeElection(racer.Ctx, racer.electionKey(), racer.leader)
	} else {
		log.Infof("new election...")
		racer.election, _ = racer.Connector.NewElection(racer.Ctx, racer.electionKey())
	}

	racer.campaign()
	racer.observe()
	log.Infof("session created.")
	return nil
}

func (racer *Elector) campaign() {
	go func() {
		log.Infof("campaign as %v...", racer.node.GetName())
		racer.ErrorChan <- racer.election.Campaign(racer.Ctx, racer.node.GetSpec().Marshal())
	}()
}

func (racer *Elector) observe() {
	go func() {
		o := racer.election.Observe(racer.Ctx)
		for leader := range o {
			racer.changeLeader(leader)
		}
	}()
}

func (racer *Elector) changeLeader(leader clientv3.GetResponse) {
	amICurrentlyTheLeader := racer.amITheLeader()
	amITheNewLeader := racer.toNominee(leader).Name == racer.node.GetName()
	racer.leader = leader
	if amITheNewLeader && amICurrentlyTheLeader {
		log.Infof("I stay the leader. Nothing to do.")
	} else if amITheNewLeader && !amICurrentlyTheLeader {
		log.Infof("promoting The Node...")
		racer.ErrorChan <- racer.node.Lead(racer.Ctx, racer.leaderNominee())
	} else if !amITheNewLeader && amICurrentlyTheLeader {
		_ = racer.node.Stonith(racer.Ctx)
		racer.Stonith()
	} else {
		racer.ErrorChan <- racer.node.Follow(racer.Ctx, racer.leaderNominee())
	}
}

func (racer *Elector) leaderNominee() node2.Spec {
	return racer.toNominee(racer.leader)
}

func (racer *Elector) amITheLeader() bool {
	return racer.leaderNominee().Name == racer.node.GetName()
}
