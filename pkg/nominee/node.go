package nominee

import (
	"context"
	"encoding/json"
)

type NodeSpecifier interface {
	GetName() string
	GetAddress() string
	GetPort() int64
	Spec() *NodeSpec
}

// Node ...
type Node interface {
	Daemon
	NodeSpecifier
	Lead(context context.Context, leader NodeSpec) error
	Follow(context context.Context, leader NodeSpec) error
	Stonith(context context.Context) error
	Stop() StopChan
}

// NodeBase ...
type NodeBase struct {
	*NodeSpec
}

// StopChan ...
type StopChan <-chan struct{}

// NodeSpec ...
type NodeSpec struct {
	ElectionKey string
	Name        string
	Address     string
	Port        int64
}

// Marshal ...
func (n *NodeSpec) Marshal() string {
	data, _ := json.Marshal(n)
	return string(data)
}

// Unmarshal ...
func Unmarshal(data []byte) (NodeSpec, error) {
	value := NodeSpec{}
	err := json.Unmarshal(data, &value)
	return value, err
}

func (n *NodeSpec) GetName() string {
	return n.Name
}

func (n *NodeSpec) GetAddress() string {
	return n.Address
}

func (n *NodeSpec) GetPort() int64 {
	return n.Port
}

func (n *NodeSpec) Spec() *NodeSpec {
	return n
}
