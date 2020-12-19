package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

type Dummy struct {
	Name    string
	Cluster string
}

func (dummy *Dummy) ServiceName() string {
	return dummy.Name
}

func (dummy *Dummy) NomineeName() string {
	return "dummy"
}
func (dummy *Dummy) NomineeAddress() string {
	return "dummy"
}

func (dummy *Dummy) ClusterName() string {
	return dummy.Cluster
}

func (dummy *Dummy) Nominee() nominee.Nominee {
	return nominee.Nominee{}
}

func (dummy *Dummy) Promote(context.Context, nominee.Nominee) error {
	return nil
}

func (dummy *Dummy) FollowNewLeader(context.Context, nominee.Nominee) error {
	return nil
}

func (dummy *Dummy) Stonith(context.Context) error {
	return nil
}

func (dummy *Dummy) ServiceStopChan() <-chan error {
	return make(chan error)
}
