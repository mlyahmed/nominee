package election

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/base"
	"github/mlyahmed.io/nominee/pkg/logger"
	"github/mlyahmed.io/nominee/pkg/node"
	"github/mlyahmed.io/nominee/pkg/stonither"
)

// Elector ...
type Elector interface {
	LeaderWatcher
	Run(node.Node) error
	stonither.Stonither
	base.Cleaner
}

// DefaultElector ...
type DefaultElector struct {
	*stonither.Base
	Managed node.Node
	Leader  *node.Spec
}

// NewElector ...
func NewElector(managed node.Node) *DefaultElector {
	elector := &DefaultElector{
		Base:    stonither.NewBase(),
		Managed: managed,
	}
	elector.listenToTheNodeStopChan()
	return elector
}

// UpdateLeader ...
func (e *DefaultElector) UpdateLeader(leader *node.Spec) error {
	amICurrentlyTheLeader := e.amITheLeader()
	amITheNewLeader := leader.Name == e.Managed.GetName()
	e.Leader = leader
	if amITheNewLeader && amICurrentlyTheLeader {
		logger.G(context.Background()).Infof("I stay the leader. Nothing to do.")
	} else if amITheNewLeader && !amICurrentlyTheLeader {
		logger.G(context.Background()).Infof("promoting The Node...")
		if err := e.Managed.Lead(e.Ctx, *e.Leader); err != nil {
			e.Stonith(context.TODO())
		}
	} else if !amITheNewLeader && amICurrentlyTheLeader {
		e.Managed.Stonith(e.Ctx)
		e.Stonith(e.Ctx)
	} else {
		if err := e.Managed.Follow(e.Ctx, *e.Leader); err != nil {
			e.Managed.Stonith(e.Ctx)
			e.Stonith(e.Ctx)
		}
	}
	return nil
}

func (e *DefaultElector) listenToTheNodeStopChan() {
	go func() {
		<-e.Managed.Done()
		e.Stonith(context.TODO())
	}()
}

func (e *DefaultElector) amITheLeader() bool {
	return e.Leader != nil && e.Leader.Name == e.Managed.GetName()
}
