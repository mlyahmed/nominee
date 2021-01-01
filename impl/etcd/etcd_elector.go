package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

// Elector ...
type Elector struct {
	*Etcd
	node     nominee.Node
	leader   clientv3.GetResponse
	election Election
}

// NewElector ...
func NewElector(config *Config) *Elector {
	logger = logrus.WithFields(logrus.Fields{"elector": "etcd", "domain": config.Domain, "cluster": config.Cluster})
	racer := Elector{Etcd: NewEtcd(config)}
	racer.failBackFn = func() error { return racer.connect(true) }
	return &racer
}

// Run ...
func (racer *Elector) Run(node nominee.Node) error {
	logger = logger.WithFields(logrus.Fields{"daemon": node.DaemonName(), "node": node.GetName()})
	logger.Infof("starting...")
	racer.node = node
	racer.forwardStopChan()
	if err := racer.connect(false); err != nil {
		return err
	}
	racer.listenToTheConnectorSession()
	logger.Infof("started.")
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

	if _, err := racer.Connector.Connect(racer.Context, racer.Config); err != nil {
		return err
	}

	if len(racer.leader.Kvs) > 0 {
		logger.Infof("resume election...")
		racer.election, _ = racer.Connector.ResumeElection(racer.Context, racer.electionKey(), racer.leader)
	} else {
		logger.Infof("new election...")
		racer.election, _ = racer.Connector.NewElection(racer.Context, racer.electionKey())
	}

	racer.conquer()
	racer.observeLeader()
	logger.Infof("session created.")
	return nil
}

func (racer *Elector) conquer() {
	go func() {
		logger.Infof("conquer as %v...", racer.node.GetName())
		racer.ErrorChan <- racer.election.Campaign(racer.Context, racer.node.Spec().Marshal())
	}()
}

func (racer *Elector) observeLeader() {
	go func() {
		observe := racer.election.Observe(racer.Context)
		for leader := range observe {
			racer.changeLeader(leader)
		}
	}()
}

func (racer *Elector) changeLeader(leader clientv3.GetResponse) {
	amICurrentlyTheLeader := racer.amITheLeader()
	amITheNewLeader := racer.toNominee(leader).Name == racer.node.GetName()
	racer.leader = leader
	if amITheNewLeader && amICurrentlyTheLeader {
		logger.Infof("I stay the leader. Nothing to do.")
	} else if amITheNewLeader && !amICurrentlyTheLeader {
		logger.Infof("promoting The Node...")
		racer.ErrorChan <- racer.node.Lead(racer.Context, racer.leaderNominee())
	} else if !amITheNewLeader && amICurrentlyTheLeader {
		_ = racer.node.Stonith(racer.Context)
		racer.Stonith()
	} else {
		racer.ErrorChan <- racer.node.Follow(racer.Context, racer.leaderNominee())
	}
}

func (racer *Elector) leaderNominee() nominee.NodeSpec {
	return racer.toNominee(racer.leader)
}

func (racer *Elector) amITheLeader() bool {
	return racer.leaderNominee().Name == racer.node.GetName()
}
