package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"testing"
)

type MockServiceRecord struct {
	LeadHits    int
	FollowHits  int
	StonithHits int
	Leader      nominee.Nominee
}

// MockService ...
type MockService struct {
	*MockServiceRecord
	t                *testing.T
	StopChan         chan struct{}
	ServiceNameFn    func() string
	NomineeNameFn    func() string
	NomineeAddressFn func() string
	NomineeFn        func() nominee.Nominee
	LeadFn           func(context.Context, nominee.Nominee) error
	FollowFn         func(context.Context, nominee.Nominee) error
	StonithFn        func(context.Context) error
	StopChanFn       func() nominee.StopChan
}

// NewMockServiceWithNominee ...
func NewMockService(t *testing.T, node *nominee.Nominee) *MockService {
	stopChan := make(chan struct{}, 1)
	return &MockService{
		MockServiceRecord: &MockServiceRecord{},
		StopChan:          stopChan,
		ServiceNameFn: func() string {
			return "mockedService"
		},
		NomineeNameFn: func() string {
			return node.Name
		},
		NomineeAddressFn: func() string {
			return node.Address
		},
		NomineeFn: func() nominee.Nominee {
			return *node
		},
		LeadFn: func(ctx context.Context, nominee nominee.Nominee) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: LeadFn function not specified.", testutils.Failed)
			return nil
		},
		FollowFn: func(ctx context.Context, n nominee.Nominee) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: FollowFn function not specified.", testutils.Failed)
			return nil
		},
		StonithFn: func(ctx context.Context) error {
			t.Fatalf("\t\t\t%s FATAL [Fail Fast]: StonithFn function not specified.", testutils.Failed)
			return nil
		},
		StopChanFn: func() nominee.StopChan {
			return stopChan
		},
	}
}

// ServiceName ...
func (mock *MockService) ServiceName() string {
	return mock.ServiceNameFn()
}

// NomineeName ...
func (mock *MockService) NomineeName() string {
	return mock.NomineeAddressFn()
}

// NomineeAddress ...
func (mock *MockService) NomineeAddress() string {
	return mock.NomineeAddressFn()
}

// Nominee ...
func (mock *MockService) Nominee() nominee.Nominee {
	return mock.NomineeFn()
}

// Lead ...
func (mock *MockService) Lead(ctx context.Context, leader nominee.Nominee) error {
	mock.LeadHits++
	mock.Leader = leader
	return mock.LeadFn(ctx, leader)
}

// Follow ...
func (mock *MockService) Follow(ctx context.Context, leader nominee.Nominee) error {
	mock.FollowHits++
	mock.Leader = leader
	return mock.FollowFn(ctx, leader)
}

// Stonith ...
func (mock *MockService) Stonith(ctx context.Context) error {
	mock.StonithHits++
	return mock.StonithFn(ctx)
}

// StopChan ...
func (mock *MockService) Stop() nominee.StopChan {
	return mock.StopChanFn()
}
