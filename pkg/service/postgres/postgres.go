package postgres

import (
	"context"
	"github.com/sirupsen/logrus"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
	"os/exec"
)

var (
	logger *logrus.Entry
)

type Postgres struct {
	stopCh chan error
	nominee nominee.Nominee
}

func NewPostgres(nominee nominee.Nominee) *Postgres {
	pg := &Postgres{
		stopCh:  make(chan error),
		nominee: nominee,
	}
	logger = logrus.WithFields(logrus.Fields{"service": pg.Name(), "node": pg.NodeName()})
	return pg
}

func (pg *Postgres) Name() string {
	return "postgres"
}

func (pg *Postgres) NodeName() string {
	return pg.nominee.Name
}

func (pg *Postgres) ClusterName() string {
	return pg.nominee.Cluster
}

func (pg *Postgres) Promote(context context.Context, _ nominee.Nominee) error {
	logger.Infof("postgres: promote to primary...\n")
	promotion := exec.CommandContext(context, "/docker-entrypoint.sh", "postgres")
	promotion.Stdout, promotion.Stderr = os.Stdout, os.Stderr
	go func() { pg.stopCh <- promotion.Run()}()
	return nil
}

func (pg *Postgres) FollowNewLeader(context.Context, nominee.Nominee) error {
	logger.Infof("postgres: following the new leader. \n")
	return nil
}

func (pg *Postgres) Stonith(context.Context) error {
	logger.Infof("postgres: stopping... \n")
	return nil
}

func (pg *Postgres) StopChan() <- chan error {
	return pg.stopCh
}
