package service

import (
	"context"
	"errors"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

var (
	ErrAbort = errors.New("service: abort")
)

type Service interface {
	ServiceName() string
	NodeName() string
	NodeAddress() string
	ClusterName() string
	Nominee() nominee.Nominee
	Promote(context context.Context, leader nominee.Nominee) error
	FollowNewLeader(context context.Context, leader nominee.Nominee) error
	Stonith(context context.Context) error
	StopChan() <-chan error
}
