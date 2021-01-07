package election

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/base"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/stonither"
	"sync"
)

// Observer ...
type Observer interface {
	LeaderWatcher
	NodesWatcher
	Observe(proxy.Proxy) error
	stonither.Stonither
	base.Cleaner
}

type DefaultObserver struct {
	*stonither.Base
	Managed   proxy.Proxy
	Leader    *node.Spec
	Followers map[string]*node.Spec
	mutex     *sync.Mutex
	updated   bool
}

func NewObserver(p proxy.Proxy) *DefaultObserver {
	observer := &DefaultObserver{
		Base:      stonither.NewBase(),
		Managed:   p,
		Leader:    nil,
		Followers: make(map[string]*node.Spec),
		mutex:     &sync.Mutex{},
		updated:   false,
	}
	observer.listenToTheProxyStopChan()
	observer.startObservationLoop()
	return observer
}

func (observer *DefaultObserver) startObservationLoop() {
	go func() {
		for {
			observer.publish()
		}
	}()
}

func (observer *DefaultObserver) publish() {
	if observer.updated {
		logger.G(context.Background()).Info("Publish to the proxy...")
		observer.mutex.Lock()
		defer observer.mutex.Unlock()

		followers := make([]*node.Spec, len(observer.Followers))
		for _, follower := range observer.Followers {
			followers = append(followers, follower)
		}
		_ = observer.Managed.Publish(observer.Leader, followers...)
		observer.updated = false
		return
	}
}

func (observer *DefaultObserver) UpdateLeader(leader *node.Spec) error {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.Leader = leader
	delete(observer.Followers, leader.ElectionKey)
	observer.updated = true
	return nil
}

func (observer *DefaultObserver) UpdateNodes(nodes []*node.Spec) error {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	for _, spec := range nodes {
		observer.Followers[spec.ElectionKey] = spec
	}
	observer.updated = true
	return nil
}

func (observer *DefaultObserver) RemoveNodes(nodes ...*node.Spec) error {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	for _, spec := range nodes {
		if observer.Leader != nil && observer.Leader.ElectionKey == spec.ElectionKey {
			observer.Leader = nil
		}
		delete(observer.Followers, spec.ElectionKey)
	}
	observer.updated = true
	return nil
}

func (observer *DefaultObserver) listenToTheProxyStopChan() {
	go func() {
		<-observer.Managed.Done()
		observer.Stonith(context.TODO())
	}()
}
