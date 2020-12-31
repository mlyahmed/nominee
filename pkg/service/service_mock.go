package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"testing"
)

// MockService ...
type MockService struct {
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
func NewMockServiceWithNominee(t *testing.T, node *nominee.Nominee) *MockService {
	stopChan := make(chan struct{}, 1)
	return &MockService{
		StopChan: stopChan,
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
			t.Fatalf("\t\t\t%s FATAL: Lead function not implemented.", testutils.Failed)
			return nil
		},
		FollowFn: func(ctx context.Context, n nominee.Nominee) error {
			t.Fatalf("\t\t\t%s FATAL: Follow function not implemented.", testutils.Failed)
			return nil
		},
		StonithFn: func(ctx context.Context) error {
			t.Fatalf("\t\t\t%s FATAL: Stonith function not implemented.", testutils.Failed)
			return nil
		},
		StopChanFn: func() nominee.StopChan {
			return stopChan
		},
	}
}

// NewMockService ...
func NewMockService(t *testing.T) *MockService {
	return NewMockServiceWithNominee(t, &nominee.Nominee{})
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
	return mock.LeadFn(ctx, leader)
}

// Follow ...
func (mock *MockService) Follow(ctx context.Context, leader nominee.Nominee) error {
	return mock.FollowFn(ctx, leader)
}

// Stonith ...
func (mock *MockService) Stonith(ctx context.Context) error {
	return mock.StonithFn(ctx)
}

// StopChan ...
func (mock *MockService) Stop() nominee.StopChan {
	return mock.StopChanFn()
}
