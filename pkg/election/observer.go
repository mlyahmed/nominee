package election

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/base"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/proxy"
	"github/mlyahmed.io/nominee/pkg/stonither"
	"sync"
	"time"
)

// Observer ...
type Observer interface {
	LeaderWatcher
	NodesWatcher
	Observe(proxy.Proxy) error
	stonither.Stonither
	base.Cleaner
}

type BasicObserver struct {
	*stonither.Basic
	Managed   proxy.Proxy
	Leader    *node.Spec
	Followers map[string]*node.Spec
	mutex     *sync.Mutex
	updated   bool
}

func NewBasicObserver(p proxy.Proxy) *BasicObserver {
	observer := &BasicObserver{
		Basic:     stonither.NewBasic(),
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

func (observer *BasicObserver) startObservationLoop() {
	go func() {
		for {
			observer.publish()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (observer *BasicObserver) publish() {
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

func (observer *BasicObserver) UpdateLeader(leader *node.Spec) error {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.Leader = leader
	delete(observer.Followers, leader.ElectionKey)
	observer.updated = true
	return nil
}

func (observer *BasicObserver) UpdateNodes(nodes []*node.Spec) error {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	for _, spec := range nodes {
		observer.Followers[spec.ElectionKey] = spec
	}
	observer.updated = true
	return nil
}

func (observer *BasicObserver) RemoveNodes(nodes ...*node.Spec) error {
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

func (observer *BasicObserver) listenToTheProxyStopChan() {
	go func() {
		<-observer.Managed.Done()
		observer.Stonith(context.TODO())
	}()
}
