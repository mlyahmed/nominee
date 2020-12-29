package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/race/etcdconfig"
	"github/mlyahmed.io/nominee/pkg/service"
)

// Racer ...
type Racer struct {
	*Etcd
	service  service.Service
	leader   clientv3.GetResponse
	election Election
}

// NewEtcdRacer ...
func NewEtcdRacer(config *etcdconfig.Config) *Racer {
	logger = logrus.WithFields(logrus.Fields{"elector": "etcd", "domain": config.Domain, "cluster": config.Cluster})
	return &Racer{Etcd: NewEtcd(config)}
}

// Run ...
func (racer *Racer) Run(service service.Service) error {
	logger = logger.WithFields(logrus.Fields{"racer": "etcd", "domain": service.ServiceName(), "nominee": service.NomineeName()})
	logger.Infof("starting...")

	racer.service = service
	racer.setUpOSSignals()

	if err := racer.newSession(); err != nil {
		return err
	}

	racer.conquer()
	racer.observeLeader()
	racer.stayTuned()

	logger.Infof("started.")
	return nil
}

func (racer *Racer) newSession() error {
	if _, err := racer.Connect(racer.ctx, racer.Config); err != nil {
		return err
	}

	if len(racer.leader.Kvs) > 0 {
		logger.Infof("resume election...")
		racer.election, _ = racer.ResumeElection(racer.ctx, racer.electionKey(), racer.leader)
	} else {
		logger.Infof("new election...")
		racer.election, _ = racer.NewElection(racer.ctx, racer.electionKey())
	}
	logger.Infof("session created.")
	return nil
}

func (racer *Racer) conquer() {
	go func() {
		logger.Infof("conquer as %v...", racer.service.NomineeName())
		racer.errorChan <- racer.election.Campaign(racer.ctx, racer.service.Nominee().Marshal())
	}()
}

func (racer *Racer) observeLeader() {
	go func() {
		observe := racer.election.Observe(racer.ctx)
		for leader := range observe {
			racer.changeLeader(leader)
		}
		logger.Debug("observation stopped.")
	}()
}

func (racer *Racer) changeLeader(leader clientv3.GetResponse) {
	amICurrentlyTheLeader := racer.amITheLeader()
	amITheNewLeader := racer.toNominee(leader).Name == racer.service.NomineeName()
	racer.leader = leader

	if amITheNewLeader && amICurrentlyTheLeader {

		logger.Infof("I stay the leader. Nothing to do.")

	} else if amITheNewLeader && !amICurrentlyTheLeader {

		logger.Infof("promoting The Service...")
		racer.errorChan <- racer.service.Lead(racer.ctx, racer.leaderNominee())

	} else if !amITheNewLeader && amICurrentlyTheLeader {

		racer.stonith()

	} else {

		racer.errorChan <- racer.service.Follow(racer.ctx, racer.leaderNominee())

	}
}

func (racer *Racer) retry() error {
	_ = racer.Etcd.retry()
	racer.conquer()
	racer.observeLeader()
	return nil
}

func (racer *Racer) stonith() {
	logger.Infof("stonithing...")

	if racer.amITheLeader() {
		logger.Infof("resign since I was leader...")
		_ = racer.service.Stonith(racer.ctx)
		_ = racer.election.Resign(racer.ctx)
	}

	racer.Etcd.stonith()
}

func (racer *Racer) leaderNominee() nominee.Nominee {
	return racer.toNominee(racer.leader)
}

func (racer *Racer) amITheLeader() bool {
	return racer.leaderNominee().Name == racer.service.NomineeName()
}

func (racer *Racer) nomineeStopChan() nominee.StopChan {
	return racer.service.StopChan()
}
