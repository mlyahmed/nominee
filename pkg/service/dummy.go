package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
)

// Dummy ...
type Dummy struct {
	nominee nominee.Nominee
}

// NewDummy ...
func NewDummy() *Dummy {
	return &Dummy{
		nominee: nominee.Nominee{
			Name: "dummy",
		},
	}
}

// ServiceName ...
func (d *Dummy) ServiceName() string {
	return "dummy"
}

// NomineeName ...
func (d *Dummy) NomineeName() string {
	return d.nominee.Name
}

// NomineeAddress ...
func (d *Dummy) NomineeAddress() string {
	return d.nominee.Address
}

// Nominee ...
func (d *Dummy) Nominee() nominee.Nominee {
	return d.nominee
}

// Lead ...
func (d *Dummy) Lead(context.Context, nominee.Nominee) error {
	return nil
}

// Follow ...
func (d *Dummy) Follow(context.Context, nominee.Nominee) error {
	return nil
}

// Stonith ...
func (d *Dummy) Stonith(context.Context) error {
	os.Exit(1)
	return nil
}

// StopChan ...
func (d *Dummy) StopChan() nominee.StopChan {
	return make(nominee.StopChan)
}
