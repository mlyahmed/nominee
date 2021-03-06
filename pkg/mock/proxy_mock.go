package mock

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/base"
	"github/mlyahmed.io/nominee/pkg/node"
)

type ProxyRecord struct {
	PublishHits int
	StonithHits int
}

type Proxy struct {
	*ProxyRecord
	Leader    *node.Spec
	Followers []*node.Spec
	PublishFn func(leader *node.Spec, followers ...*node.Spec) error
	doneChan  chan struct{}
}

func NewProxy() *Proxy {
	return &Proxy{
		ProxyRecord: &ProxyRecord{},
		Followers:   make([]*node.Spec, 0),
		PublishFn: func(leader *node.Spec, followers ...*node.Spec) error {
			return nil
		},
		doneChan: make(chan struct{}),
	}
}

func (p *Proxy) Publish(leader *node.Spec, followers ...*node.Spec) error {
	p.PublishHits++
	p.Leader = leader
	p.Followers = followers
	return p.PublishFn(leader, followers...)
}

func (p *Proxy) Stonith(context.Context) {
	p.StonithHits++
	close(p.doneChan)
}

func (p *Proxy) Done() base.DoneChan {
	return p.doneChan
}
