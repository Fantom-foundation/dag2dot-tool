package types

import (
	"strconv"

	"github.com/Fantom-foundation/dag2dot-tool/rpc"
)

// A node to query to event data from
type EventNode struct {
	rpc.Event
	NodeName string
	NodeGroup string
}

func NewEventNode(ev *rpc.Event) *EventNode {
	trx := ""
	if len(ev.Transactions) > 0 {
		trx = " (trxs: "+strconv.FormatInt(int64(len(ev.Transactions)), 10)+")"
	}
	nodeName := strconv.FormatInt(ev.Epoch, 10)+"-"+strconv.FormatInt(ev.Lamport, 10)+"-"+ev.Hash[len(ev.Hash)-8:] +
		"\n"+strconv.FormatInt(ev.Frame, 10)+"-"+strconv.FormatInt(ev.Seq, 10)+trx

	return &EventNode{
		Event: *ev,
		NodeName:    nodeName,
		NodeGroup:	 "host-"+ strconv.FormatInt(ev.Creator, 10),
	}
}

func (n EventNode) GetId() string {
	return strconv.FormatInt(n.Creator, 10)
}

