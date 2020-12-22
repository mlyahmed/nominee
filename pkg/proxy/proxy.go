package proxy

import "github/mlyahmed.io/nominee/pkg/nominee"

type Proxy interface {
	PushNominees(nominees ...nominee.Nominee) error
	PushLeader(leader nominee.Nominee) error
	RemoveNominee(electionKey string) error
}
