package race

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/service"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type EtcdRacer struct {
	*Etcd
	service service.Service
	leader  clientv3.GetResponse
}

func NewEtcdRacer(endpoints []string) Racer {
	subContext, subCancel := context.WithCancel(context.Background())
	logger = logrus.WithFields(logrus.Fields{"elector": "etcd"})
	return &EtcdRacer{
		Etcd: &Etcd{
			endpoints: endpoints,
			ctx:       subContext,
			cancel:    subCancel,
			errorChan: make(chan error),
			stopChan:  make(chan error),
		},
	}
}

func (racer *EtcdRacer) Run(service service.Service) error {
	logger = logger.WithFields(logrus.Fields{"racer": "etcd", "service": service.ServiceName(), "nominee": service.NomineeName()})
	logger.Infof("starting...")

	racer.service = service
	racer.domain = service.ServiceName()
	racer.cluster = service.ClusterName()

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

func (racer *EtcdRacer) newSession() error {
	if err := racer.Etcd.newSession(); err != nil {
		return err
	}

	if len(racer.leader.Kvs) > 0 {
		logger.Infof("resume election...")
		racer.election = concurrency.ResumeElection(racer.session, racer.electionKey(), string(racer.leader.Kvs[0].Key), racer.leader.Kvs[0].CreateRevision)
	} else {
		logger.Infof("new election...")
		racer.election = concurrency.NewElection(racer.session, racer.electionKey())
	}
	logger.Infof("session created.")
	return nil
}

func (racer *EtcdRacer) conquer() {
	go func() {
		logger.Infof("conquer as %v...", racer.service.NomineeName())
		racer.errorChan <- racer.election.Campaign(racer.ctx, racer.service.Nominee().Marshal())
	}()
}

func (racer *EtcdRacer) observeLeader() {
	go func() {
		observe := racer.election.Observe(racer.ctx)
		for leader := range observe {
			racer.changeLeader(leader)
		}
		logger.Debug("observation stopped.")
	}()
}

func (racer *EtcdRacer) changeLeader(leader clientv3.GetResponse) {
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

func (racer *EtcdRacer) retry() error {
	_ = racer.Etcd.retry()
	racer.conquer()
	racer.observeLeader()
	return nil
}

func (racer *EtcdRacer) stonith() {
	logger.Infof("stonithing...")

	if racer.amITheLeader() {
		logger.Infof("resign since I was leader...")
		_ = racer.service.Stonith(racer.ctx)
		_ = racer.election.Resign(racer.ctx)
	}

	racer.Etcd.stonith()
}

func (racer *EtcdRacer) leaderNominee() nominee.Nominee {
	return racer.toNominee(racer.leader)
}

func (racer *EtcdRacer) amITheLeader() bool {
	return racer.leaderNominee().Name == racer.service.NomineeName()
}

func (racer *EtcdRacer) nomineeStopChan() nominee.StopChan {
	return racer.service.StopChan()
}
