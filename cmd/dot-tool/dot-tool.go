package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Fantom-foundation/go-opera/ftmclient"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang-collections/collections/stack"

	"github.com/Fantom-foundation/dag2dot-tool/dot"
	"github.com/Fantom-foundation/dag2dot-tool/types"
)

// The main entry point of the dagtool.

const (
	colorRoot    = "#FFFF00"
	colorNewRoot = "#AAAA00"
	colorOldRoot = "#888888"
)

var (
	LatestSealedEpoch = big.NewInt(-1)
)

// configs
type Config struct {
	RPCHost    string
	RPCPort    int
	OutPath    string
	LvlLimit   int
	OnlyEpoch  bool
	RenderFile bool
}

// main function
func main() {
	var cfg Config
	var mode string

	flag.StringVar(&cfg.RPCHost, "host", "localhost", "Host for RPC requests")
	flag.IntVar(&cfg.RPCPort, "port", 18545, "Port for RPC requests")
	flag.IntVar(&cfg.LvlLimit, "limit", 0, "DAG level limit")
	flag.StringVar(&cfg.OutPath, "out", "", "Path of directory for save DOT files")
	flag.StringVar(&mode, "mode", "root", "Mode:\nroot - single shot to every root node changes\nepoch - single shot to every epoch")
	flag.BoolVar(&cfg.RenderFile, "render", true, "Render:\n true - render dot file to png image\n false - no rendering")
	flag.Parse()

	if cfg.OutPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg.OnlyEpoch = mode == "epoch"

	ProcessLoop(cfg)
}

func ProcessLoop(cfg Config) {

	url := fmt.Sprintf("http://%s:%d/", cfg.RPCHost, cfg.RPCPort)
	conn, err := rpc.Dial(url)
	if err != nil {
		log.Panicf("Can not connect RPC: %s\n", err)
	}
	r := ftmclient.NewClient(conn)
	ctx := context.TODO()

	processedTop := make(map[hash.Event]bool)

	var prevGraphData *types.GraphData
	var prevGraph *dot.Graph
	var prevEpoch idx.Epoch

mainLoop:
	for {
		graphName := "DAG" + strconv.FormatInt(time.Now().UnixNano(), 10)

		subGraphs := make(map[string]*dot.SubGraph)
		extEdges := make([]*dot.Edge, 0)
		graphData := &types.GraphData{}

		// Get top events
		top, err := r.GetHeads(ctx, LatestSealedEpoch)
		if err != nil {
			log.Panicf("Can not get top events: %s\n", err)
		}
		if len(top) == 0 {
			log.Printf("No data for loop %s\n", graphName)
			time.Sleep(1 * time.Second)
			continue mainLoop
		}

		nodes := make(map[hash.Event]*types.EventNode)
		inGraph := make(map[string]*dot.Node)

		hashStack := stack.New()

		var startLevel idx.Event
		var curEpoch idx.Epoch

		newEpoch := false

		for _, h := range top {
			if processedTop[h] {
				time.Sleep(100 * time.Millisecond)
				continue mainLoop
			}
			processedTop[h] = true

			head, err := r.GetEvent(ctx, h)
			if err != nil {
				log.Panicf("Can not get head: %s\n", err)
			}
			curEpoch = head.Epoch()

			if head.Epoch() != prevEpoch {
				newEpoch = true
			}

			startLevel = head.Seq()

			p := types.NewEventNode(head)
			nodes[h] = p

			n := dot.NewNode(p.NodeName)
			graphData.AddNode(n)
			// TODO: restore isRoot attribute
			/*
				if p.IsRoot {
					n.Set("style", "filled")
					n.Set("fillcolor", colorRoot)
				}*/

			sg, ok := subGraphs[p.NodeGroup]
			if !ok {
				idx := len(subGraphs)
				sg = dot.NewSubgraph("cluster" + strconv.FormatInt(int64(idx), 10))
				sg.Set("style", "dotted")
				sg.Set("label", p.NodeGroup)
				id, _ := strconv.ParseInt(p.GetId(), 16, 64)
				sg.Set("sortv", strconv.FormatInt(id, 10))
				subGraphs[p.NodeGroup] = sg

				pseudoNode := dot.NewNode(p.NodeGroup)
				graphData.AddNode(pseudoNode)
				pseudoNode.Set("style", "invis")
				pseudoNode.Set("width", "0")
				sg.AddNode(pseudoNode)
				inGraph[p.NodeGroup] = pseudoNode
			}
			n.Set("shape", "tripleoctagon")

			sg.AddNode(n)

			inGraph[p.NodeName] = n

			hashStack.Push(h)
		}

		log.Printf("Start loop %s\n", graphName)

		// log.Printf("DBG1\n", )

		processed := make(map[hash.Event]bool)

		for hashStack.Len() > 0 {
			h := hashStack.Pop().(hash.Event)
			if processed[h] {
				// Skip already processed hodes
				continue
			}
			processed[h] = true

			// Get current node
			node, present := nodes[h]
			if !present {
				head, err := r.GetEvent(ctx, h)
				if err != nil {
					log.Panicf("Can not get head: %s\n", err)
				}

				node = types.NewEventNode(head)
			}

			if cfg.LvlLimit > 0 && int(startLevel-node.Seq()) > cfg.LvlLimit {
				log.Println("Finish DAG by limit")
				break
			}
			mainNode := inGraph[node.NodeName]

			// For all parents
			for _, parent := range node.Parents() {
				// Get parent node
				p, present := nodes[parent]
				if !present {
					head, err := r.GetEvent(ctx, parent)
					if err != nil {
						log.Panicf("Can not get head: %s\n", err)
					}

					p = types.NewEventNode(head)

					// Save to nodes cache
					nodes[parent] = p
				}

				// Add parent node to graph
				n, ok := inGraph[p.NodeName]
				if !ok {
					n = dot.NewNode(p.NodeName)
					graphData.AddNode(n)
					// TODO: restore isRoot attribute
					/*
						if p.IsRoot {
							n.Set("style", "filled")
							n.Set("fillcolor", colorRoot)
						}
					*/
					sg, ok := subGraphs[p.NodeGroup]
					if !ok {
						idx := len(subGraphs)
						sg = dot.NewSubgraph("cluster" + strconv.FormatInt(int64(idx), 10))
						sg.Set("label", p.NodeGroup)
						sg.Set("style", "dotted")
						id, _ := strconv.ParseInt(p.GetId(), 16, 64)
						sg.Set("sortv", strconv.FormatInt(id, 10))
						subGraphs[p.NodeGroup] = sg

						pseudoNode := dot.NewNode(p.NodeGroup)
						graphData.AddNode(pseudoNode)
						pseudoNode.Set("style", "invis")
						pseudoNode.Set("width", "0")
						sg.AddNode(pseudoNode)
						inGraph[p.NodeGroup] = pseudoNode
					}
					sg.AddNode(n)
					inGraph[p.NodeName] = n
				}
				// Add edge from main node to parent
				e := dot.NewEdge(mainNode, n)
				graphData.AddEdge(e)
				e.Set("constraint", "true")
				if node.NodeGroup == p.NodeGroup {
					sg, _ := subGraphs[p.NodeGroup]
					sg.AddEdge(e)
				} else {
					extEdges = append(extEdges, e)
				}

				// Add parent node for processing on next loop
				hashStack.Push(parent)
			}
		}

		// Create graph
		g := dot.NewGraph(graphName)
		// set attribs to local
		g.Set("clusterrank", "local")
		g.Set("compound", "true")
		g.Set("newrank", "true")
		g.Set("ranksep", "0.05")

		// Sort subgraphs names
		subGraphsNames := make([]string, 0, len(subGraphs))
		for sgName, _ := range subGraphs {
			subGraphsNames = append(subGraphsNames, sgName)
		}
		sort.Strings(subGraphsNames)

		// FIXED: dot program renders subgraphs not in the ordering that specified
		//   so we introduce pseudo nodes and edges to work around
		rankAttrib := make([]string, 1, 1)
		rankAttrib[0] = "\"" + strings.Join(subGraphsNames, `" -> "`) + "\" [style = invis, constraint = true];"
		g.SameRank(rankAttrib)

		// Add subgraphs in graph with sort order
		for _, subName := range subGraphsNames {
			g.AddSubgraph(subGraphs[subName])
		}

		// Add external edges in graph
		for _, edge := range extEdges {
			g.AddEdge(edge)
		}

		// Compare graphs elements and mark red changes
		graphData.MarkChanges(prevGraphData, "red", "2.5", colorRoot, colorNewRoot, colorOldRoot)
		prevGraphData = graphData

		if cfg.OnlyEpoch && newEpoch && prevGraph != nil {
			g = prevGraph
			log.Println("New epoch out")
		}

		if !cfg.OnlyEpoch || prevEpoch != 0 {
			flushToFile(&cfg, prevEpoch, g)
		}

		prevGraph = g
		if newEpoch {
			prevEpoch = curEpoch
		}
		log.Println("Capture loop done")
	}
}
