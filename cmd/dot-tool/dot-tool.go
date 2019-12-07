package main

import (
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/windler/dotgraph/renderer"

	"github.com/Fantom-foundation/dag2dot-tool/dot"
	"github.com/Fantom-foundation/dag2dot-tool/rpc"
	"github.com/Fantom-foundation/dag2dot-tool/types"
)

// The main entry point of the dagtool.

const (
	colorRoot    = "#FFFF00"
	colorNewRoot = "#AAAA00"
	colorOldRoot = "#888888"
)

// configs
type Config struct {
	RPCHost   string
	RPCPort   int
	OutPath   string
	LvlLimit  int
	OnlyEpoch bool
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
	r := rpc.NewRPC(cfg.RPCHost, cfg.RPCPort)

	processedTop := make(map[string]bool)

	var prevGraphData *types.GraphData
	var prevGraph *dot.Graph
	var prevEpoch int64

mainLoop:
	for {
		graphName := "DAG" + strconv.FormatInt(time.Now().UnixNano(), 10)

		subGraphs := make(map[string]*dot.SubGraph)
		extEdges := make([]*dot.Edge, 0)
		graphData := &types.GraphData{}

		// Get top events
		top, err := r.GetTopHeads()
		if err != nil {
			log.Panicf("Can not get top events: %s\n", err)
		}

		// log.Printf("TOP: %+v\n", top)

		nodes := make(map[string]*types.EventNode)
		inGraph := make(map[string]*dot.Node)

		hashStack := stack.New()

		var startLevel int64
		var curEpoch int64

		newEpoch := false

		if len(top.Result) == 0 {
			log.Printf("No data for loop %s\n", graphName)
			time.Sleep(1 * time.Second)
			newEpoch = true
			continue mainLoop
		}

		for _, h := range top.Result {
			if processedTop[h] {
				time.Sleep(100 * time.Millisecond)
				continue mainLoop
			}
			processedTop[h] = true

			head, err := r.GetEvent(h)
			if err != nil {
				log.Panicf("Can not get head: %s\n", err)
			}
			curEpoch = head.Epoch

			if head.Epoch != prevEpoch {
				newEpoch = true
			}

			startLevel = head.Seq

			// log.Printf("TOP head: %+v\n", head)

			p := types.NewEventNode(head)
			nodes[h] = p

			n := dot.NewNode(p.NodeName)
			graphData.AddNode(n)
			if p.IsRoot {
				n.Set("style", "filled")
				n.Set("fillcolor", colorRoot)
			}

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

		processed := make(map[string]bool)

		for hashStack.Len() > 0 {
			hash := hashStack.Pop().(string)
			if processed[hash] {
				// Skip already processed hodes
				continue
			}
			processed[hash] = true

			// log.Printf("DBG2: %s\n", hash)
			// Get current node
			node, present := nodes[hash]
			if !present {
				head, err := r.GetEvent(hash)
				if err != nil {
					log.Panicf("Can not get head: %s\n", err)
				}

				node = types.NewEventNode(head)
			}

			if cfg.LvlLimit > 0 && (startLevel-node.Seq) > int64(cfg.LvlLimit) {
				log.Println("Finish DAG by limit")
				break
			}
			mainNode := inGraph[node.NodeName]

			// For all parents
			for _, parent := range node.Parents {
				// log.Println("DBG3")
				// Get parent node
				p, present := nodes[parent]
				if !present {
					head, err := r.GetEvent(parent)
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
					if p.IsRoot {
						n.Set("style", "filled")
						n.Set("fillcolor", colorRoot)
					}

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
		rankAttrib[0]= "\"" + strings.Join(subGraphsNames,`" -> "`) + "\" [style = invis, constraint = true];"
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
			// Save dot file
			fileName := strings.TrimRight(cfg.OutPath, "/") + "/" + graphName + ".dot"
			if cfg.OnlyEpoch {
				fileName = strings.TrimRight(cfg.OutPath, "/") + "/" + "DAG-EPOCH-" + strconv.FormatInt(prevEpoch, 10) + ".dot"
			}
			fl, err := os.Create(fileName)
			if err != nil {
				log.Panicf("Can not create file '%s': %s\n", fileName, err)
			}
			_, err = fl.WriteString(g.String())
			if err != nil {
				log.Panicf("Can not write data to file '%s': %s\n", fileName, err)
			}
			_ = fl.Close()

			// Save png file
			if cfg.RenderFile {
				pngFileName := strings.TrimRight(cfg.OutPath, "/") + "/" + graphName + ".png"
				if cfg.OnlyEpoch {
					pngFileName = strings.TrimRight(cfg.OutPath, "/") + "/" + "DAG-EPOCH-" + strconv.FormatInt(prevEpoch, 10) + ".png"
				}
				r := &renderer.PNGRenderer{
					OutputFile: pngFileName + ".tmp",
				}
				r.Render(g.String())
				// Remove temporary file
				_ = os.Remove(pngFileName + ".tmp.dot")
				// Move tmp file to png
				_ = os.Rename(pngFileName+".tmp", pngFileName)
			}
		}

		prevGraph = g
		if newEpoch {
			prevEpoch = curEpoch
		}
		log.Println("Capture loop done")
	}
}
