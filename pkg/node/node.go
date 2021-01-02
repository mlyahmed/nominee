package node

import (
	"context"
	"encoding/json"
)

type Specifier interface {
	GetDaemonName() string
	GetName() string
	GetAddress() string
	GetPort() int64
	GetSpec() *Spec
}

// Node ...
type Node interface {
	Specifier
	Lead(context context.Context, leader Spec) error
	Follow(context context.Context, leader Spec) error
	Stonith(context context.Context) error
	Stop() StopChan
}

// Spec ...
type Spec struct {
	ElectionKey string
	Name        string
	Address     string
	Port        int64
}

// StopChan ...
type StopChan <-chan struct{}

// Marshal ...
func (n *Spec) Marshal() string {
	data, _ := json.Marshal(n)
	return string(data)
}

// Unmarshal ...
func Unmarshal(data []byte) (Spec, error) {
	value := Spec{}
	err := json.Unmarshal(data, &value)
	return value, err
}

func (n *Spec) GetName() string {
	return n.Name
}

func (n *Spec) GetAddress() string {
	return n.Address
}

func (n *Spec) GetPort() int64 {
	return n.Port
}

func (n *Spec) GetSpec() *Spec {
	return n
}