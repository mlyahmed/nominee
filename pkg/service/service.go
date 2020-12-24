package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
)

// Service ...
type Service interface {
	ServiceName() string
	NomineeName() string
	NomineeAddress() string
	Nominee() nominee.Nominee
	Lead(context context.Context, leader nominee.Nominee) error
	Follow(context context.Context, leader nominee.Nominee) error
	Stonith(context context.Context) error
	StopChan() nominee.StopChan
}
