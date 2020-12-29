package service

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/nominee"
	"os"
)

type Dummy struct {
	nominee nominee.Nominee
}

func NewDummy() *Dummy {
	return &Dummy{
		nominee: nominee.Nominee{
			Name: "dummy",
		},
	}
}

func (d *Dummy) ServiceName() string {
	return "dummy"
}

func (d *Dummy) NomineeName() string {
	return d.nominee.Name
}

func (d *Dummy) NomineeAddress() string {
	return d.nominee.Address
}

func (d *Dummy) Nominee() nominee.Nominee {
	return d.nominee
}

func (d *Dummy) Lead(context.Context, nominee.Nominee) error {

	return nil
}

func (d *Dummy) Follow(context.Context, nominee.Nominee) error {
	return nil
}

func (d *Dummy) Stonith(context.Context) error {
	os.Exit(1)
	return nil
}

func (d *Dummy) StopChan() nominee.StopChan {
	return make(nominee.StopChan)
}
