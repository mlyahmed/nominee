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
	Name() string
	NodeName() string
	ClusterName() string
	Promote(context context.Context, leader nominee.Nominee) error
	FollowNewLeader(context context.Context, leader nominee.Nominee) error
	Stonith(context context.Context) error
	StopChan() <- chan error
}
