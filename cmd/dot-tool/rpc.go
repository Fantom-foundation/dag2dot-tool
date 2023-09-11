package main

import (
	"context"

	"github.com/Fantom-foundation/go-opera/ftmclient"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
)

type RpcStoreAdapter struct {
	*ftmclient.Client
}

func (r *RpcStoreAdapter) HasEvent(id hash.Event) bool {
	// TODO: cache it
	e := r.GetEvent(id)
	return (e != nil)
}

func (r *RpcStoreAdapter) GetEvent(id hash.Event) dag.Event {
	ctx := context.TODO()
	e, err := r.Client.GetEvent(ctx, id)
	if err != nil {
		// TODO: log it
		return nil
	}

	return e
}
