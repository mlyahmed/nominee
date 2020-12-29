package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

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

func NewMockService() *MockService {
	return &MockService{
		ServiceNameFn: func() string {
			return "mockedService"
		},
		NomineeNameFn: func() string {
			return "MockedNominee"
		},
		NomineeAddressFn: func() string {
			return "unknown"
		},
		NomineeFn: func() nominee.Nominee {
			return nominee.Nominee{}
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

func (mock *MockService) ServiceName() string {
	return mock.ServiceNameFn()
}

func (mock *MockService) NomineeName() string {
	return mock.NomineeAddressFn()
}

func (mock *MockService) NomineeAddress() string {
	return mock.NomineeAddressFn()
}

func (mock *MockService) Nominee() nominee.Nominee {
	return mock.NomineeFn()
}

func (mock *MockService) Lead(ctx context.Context, leader nominee.Nominee) error {
	return mock.LeadFn(ctx, leader)
}

func (mock *MockService) Follow(ctx context.Context, leader nominee.Nominee) error {
	return mock.FollowFn(ctx, leader)
}

func (mock *MockService) Stonith(ctx context.Context) error {
	return mock.StonithFn(ctx)
}
func (mock *MockService) StopChan() nominee.StopChan {
	return mock.StopChanFn()
}
