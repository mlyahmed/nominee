package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

// MockService ...
type MockService struct {
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
func NewMockServiceWithNominee(node *nominee.Nominee) *MockService {
	return &MockService{
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
		LeadFn: func(context.Context, nominee.Nominee) error {
			return nil
		},
		FollowFn: func(context.Context, nominee.Nominee) error {
			return nil
		},
		StonithFn: func(context.Context) error {
			return nil
		},
		StopChanFn: func() nominee.StopChan {
			return nil
		},
	}
}

// NewMockService ...
func NewMockService() *MockService {
	return NewMockServiceWithNominee(&nominee.Nominee{})
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
func (mock *MockService) StopChan() nominee.StopChan {
	return mock.StopChanFn()
}
