package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/Fantom-foundation/go-opera/ftmclient"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/vecmt"
	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagordering"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/golang-collections/collections/stack"

	"github.com/Fantom-foundation/dag2dot-tool/dot"
)

// readCurrentEpochDag reads epoch's events into inmem dot.Graph
func readCurrentEpochDag(rpc *ftmclient.Client, cfg integration.Configs) *dot.Graph {
	// 0. Set gossip data:

	cdb := abft.NewMemStore()
	defer cdb.Close()

	dagIndexer := vecmt.NewIndex(panics("Vector clock"), cfg.VectorClock)
	orderer := abft.NewOrderer(
		cdb,
		&RpcStoreAdapter{rpc},
		vecmt2dagidx.Wrap(dagIndexer),
		panics("Lachesis"),
		cfg.Lachesis)
	err := orderer.Bootstrap(abft.OrdererCallbacks{})
	if err != nil {
		panic(err)
	}

	// 1. Set dot.Graph data:

	graph := dot.NewGraph("DOT")
	graph.Set("clusterrank", "local")
	graph.Set("newrank", "true")
	graph.Set("ranksep", "0.05")
	graph.Set("compound", "true")
	graph.SetGlobalEdgeAttr("constraint", "true")
	var (
		clusters  idx.ValidatorID
		subGraphs map[idx.ValidatorID]*dot.SubGraph
		emitters  []string
		nodes     map[hash.Event]*dot.Node
	)

	// 2. Set event processor data:

	var (
		epoch     idx.Epoch
		processed map[hash.Event]dag.Event
	)

	finishCurrentEpoch := func(epoch idx.Epoch) {
		for f := idx.Frame(0); f <= cdb.GetLastDecidedFrame(); f++ {
			rr := cdb.GetFrameRoots(f)
			for _, r := range rr {
				n := nodes[r.ID]
				markAsRoot(n)
			}
		}

		blockFrom, err := rpc.GetEpochBlock(context.TODO(), epoch)
		if err != nil {
			panic(err) // TODO: try again
		}
		blockTo, err := rpc.GetEpochBlock(context.TODO(), epoch+1)
		if err != nil {
			panic(err) // TODO: try again
		}
		if blockTo == 0 {
			blockTo = idx.Block(math.MaxUint64)
		}

		for b := blockFrom + 1; b <= blockTo; b++ {
			block, err := rpc.BlockByNumber(context.TODO(), big.NewInt(int64(b)))
			if err != nil {
				panic(err) // TODO: try again
			}
			if block == nil {
				break
			}
			n := nodes[hash.Event(block.Hash())]
			markAsAtropos(n)
		}

		// NOTE: github.com/tmc/dot renders subgraphs not in the ordering that specified
		//   so we introduce pseudo nodes and edges to work around
		graph.SameRank([]string{
			"\"" + strings.Join(emitters, `" -> "`) + "\" [style = invis, constraint = true];",
		})
	}

	resetToNewEpoch := func(epoch idx.Epoch) {
		// ApplyGenesis()
		cdb.SetEpochState(&abft.EpochState{
			Epoch: epoch,
		})
		cdb.SetLastDecidedState(&abft.LastDecidedState{
			LastDecidedFrame: abft.FirstFrame - 1,
		})

		vv, err := rpc.GetValidators(context.TODO(), epoch)
		if err != nil {
			panic(err) // TODO: try again
		}
		validators := parseValidatorProfiles(vv)
		sortedIDs := make([]idx.ValidatorID, validators.Len())
		copy(sortedIDs, validators.IDs())
		sort.Slice(sortedIDs, func(i, j int) bool {
			return sortedIDs[i] < sortedIDs[j]
		})
		subGraphs = make(map[idx.ValidatorID]*dot.SubGraph, len(sortedIDs))
		emitters = make([]string, 0, len(sortedIDs))
		var maxID idx.ValidatorID
		for _, v := range sortedIDs {
			if maxID < v {
				maxID = v
			}
			emitter := fmt.Sprintf("emitter-%d", clusters+v)
			emitters = append(emitters, emitter)
			sg := dot.NewSubgraph(fmt.Sprintf("cluster%d", clusters+v))
			sg.Set("label", fmt.Sprintf("emitter-%d", v))
			sg.Set("sortv", fmt.Sprintf("%d", clusters+v))
			sg.Set("style", "dotted")
			subGraphs[v] = sg
			graph.AddSubgraph(sg)

			pseudoNode := dot.NewNode(emitter)
			pseudoNode.Set("style", "invis")
			pseudoNode.Set("width", "0")
			sg.AddNode(pseudoNode)
		}
		clusters += maxID

		nodes = make(map[hash.Event]*dot.Node)
		processed = make(map[hash.Event]dag.Event, 1000)
		err = orderer.Reset(epoch, validators)
		if err != nil {
			panic(err)
		}
		dagIndexer.Reset(validators, memorydb.New(), func(id hash.Event) dag.Event {
			e, err := rpc.GetEvent(context.TODO(), id)
			if err != nil {
				panic(err) // TODO: try again
			}
			return e
		})
	}

	buffer := dagordering.New(
		cfg.Opera.Protocol.DagProcessor.EventsBufferLimit,
		dagordering.Callback{
			Process: func(e dag.Event) error {
				processed[e.ID()] = e
				err = dagIndexer.Add(e)
				if err != nil {
					panic(err)
				}
				dagIndexer.Flush()
				orderer.Process(e)

				name := e.ID().String()
				n := dot.NewNode(name)
				sg := subGraphs[e.Creator()]
				sg.AddNode(n)
				nodes[e.ID()] = n

				for _, h := range e.Parents() {
					p := nodes[h]
					ref := dot.NewEdge(n, p)
					if processed[h].Creator() == e.Creator() {
						sg.AddEdge(ref)
					} else {
						graph.AddEdge(ref)
					}
				}

				return nil
			},
			Released: func(e dag.Event, peer string, err error) {
				if err != nil {
					panic(err)
				}
			},
			Get: func(id hash.Event) dag.Event {
				return processed[id]
			},
			Exists: func(id hash.Event) bool {
				_, ok := processed[id]
				return ok
			},
		})

	// 3. Iterate over events:
	top, err := rpc.GetHeads(context.TODO(), PendingEpoch)
	if err != nil {
		log.Panicf("Can not get top events: %s\n", err)
	}
	if len(top) == 0 {
		log.Printf("No epoch heads\n")
		time.Sleep(1 * time.Second)
		return nil
	}
	hashStack := stack.New()
	for _, h := range top {
		hashStack.Push(h)
	}
	for hashStack.Len() > 0 {
		h := hashStack.Pop().(hash.Event)
		e, err := rpc.GetEvent(context.TODO(), h)
		if err != nil {
			log.Panicf("Can not get event: %s\n", err)
		}

		if epoch == 0 {
			epoch = e.Epoch()
			resetToNewEpoch(epoch)
		}

		buffer.PushEvent(e, "")

		for _, parent := range e.Parents() {
			hashStack.Push(parent)
		}
	}
	finishCurrentEpoch(epoch)

	// 4. Result:
	return graph
}

func panics(name string) func(error) {
	return func(err error) {
		log.Fatalf(err.Error())
	}
}

func markAsRoot(n *dot.Node) {
	// n.setAttr("xlabel", "root")
	n.Set("style", "filled")
	n.Set("fillcolor", "#FFFF00")
}

func markAsAtropos(n *dot.Node) {
	// n.setAttr("xlabel", "atropos")
	n.Set("style", "filled")
	n.Set("fillcolor", "#FF0000")
}

func parseValidatorProfiles(vv inter.ValidatorProfiles) *pos.Validators {
	b := pos.NewBuilder()
	for id, v := range vv {
		w := pos.Weight(v.Weight.Uint64())
		b.Set(id, w)
	}
	return b.Build()
}
