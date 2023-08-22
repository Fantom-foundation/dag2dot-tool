package types

import (
	"fmt"

	"github.com/Fantom-foundation/go-opera/inter"
)

// A node to query to event data from
type EventNode struct {
	inter.EventI
	NodeName  string
	NodeGroup string
}

func NewEventNode(ev inter.EventI) *EventNode {
	return &EventNode{
		EventI:    ev,
		NodeName:  fmt.Sprintf("%s\n%d-%d", ev.ID().String(), ev.Frame(), ev.Seq()),
		NodeGroup: fmt.Sprintf("host-%d", ev.Creator()),
	}
}

func (n EventNode) GetId() string {
	return fmt.Sprintf("%d", n.Creator())
}
