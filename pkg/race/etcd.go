package race

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/signals"
	"go.etcd.io/etcd/clientv3/concurrency"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Etcd struct {
	domain    string
	cluster   string
	endpoints []string
	ctx       context.Context
	cancel    func()
	client    *clientv3.Client
	session   *concurrency.Session
	election  *concurrency.Election
	errorChan chan error
	stopChan  chan error
	stopped   bool
}

var (
	logger *logrus.Entry
)

func NewEtcd(config *EtcdConfig) *Etcd {
	ctx, cancel := context.WithCancel(context.Background())
	return &Etcd{
		endpoints: strings.Split(config.endpoints, ","),
		ctx:       ctx,
		cancel:    cancel,
		errorChan: make(chan error),
		stopChan:  make(chan error),
		cluster:   config.Cluster,
		domain:    config.Domain,
	}
}

func (etcd *Etcd) Cleanup() {
	if etcd.client != nil {
		_ = etcd.client.Close()
	}

	if etcd.session != nil {
		_ = etcd.session.Close()
	}
}

func (etcd *Etcd) StopChan() nominee.StopChan {
	return etcd.stopChan
}

func (etcd *Etcd) setUpOSSignals() {
	listener := make(chan os.Signal, len(signals.ShutdownSignals))
	signal.Notify(listener, signals.ShutdownSignals...)
	go func() {
		<-listener
		etcd.stonith()
		<-listener
		os.Exit(1)
	}()
}

func (etcd *Etcd) newSession() error {
	logger.Infof("create new session. Endpoints %s", etcd.endpoints)
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   etcd.endpoints,
		DialTimeout: 1 * time.Second,
	})
	if err != nil {
		return err
	}
	etcd.client = client

	session, err := concurrency.NewSession(etcd.client, concurrency.WithTTL(1), concurrency.WithContext(etcd.ctx))
	if err != nil {
		return err
	}

	etcd.session = session

	return nil
}

func (etcd *Etcd) stayTuned() {
	go func() {
		for {
			select {
			case err := <-etcd.nomineeStopChan():
				logger.Warningf("nominee stopped because of %s.", err)
				etcd.stonith()
			case err := <-etcd.errorChan:
				if err != nil {
					if errors.Cause(err) == context.Canceled {
						logger.Errorf("receive cancel error. Proceed to STOP !")
						etcd.stonith()
						return
					}
					logger.Warnf("ignored error %s", err)
				}
			case <-etcd.session.Done():
				if etcd.stopped {
					return
				}
				logger.Infof("session closed. Retrying...")
				_ = etcd.retry()
			}
		}
	}()
}

func (etcd *Etcd) retry() error {
	logger.Infof("retrying...")

	etcd.cancel()
	etcd.ctx, etcd.cancel = context.WithCancel(context.Background())

	for {
		if err := etcd.newSession(); err != nil {
			logger.Errorf("error when retry new session, %s", err)
			time.Sleep(time.Second * 2)
		} else {
			logger.Info("new session created.")
			break
		}
	}

	return nil
}

func (etcd *Etcd) stonith() {
	etcd.cancel()
	close(etcd.stopChan)
	etcd.stopped = true
}

func (etcd *Etcd) toNominee(response clientv3.GetResponse) nominee.Nominee {
	var value nominee.Nominee
	if len(response.Kvs) > 0 {
		value, _ = nominee.Unmarshal(response.Kvs[0].Value)
		value.ElectionKey = string(response.Kvs[0].Key)
	}
	return value
}

func (etcd *Etcd) electionKey() string {
	return fmt.Sprintf("nominee/domain/%s/cluster/%s", etcd.domain, etcd.cluster)
}

func (etcd *Etcd) nomineeStopChan() nominee.StopChan {
	return make(chan error)
}
