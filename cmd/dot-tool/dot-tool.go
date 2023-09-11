package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/Fantom-foundation/go-opera/ftmclient"
	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/vecmt"
	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/dag2dot-tool/dot"
	"github.com/Fantom-foundation/dag2dot-tool/types"
)

// The main entry point of the dagtool.

const (
	colorAtropos    = "#AA0000"
	colorNewAtropos = "#FF0000"
	colorRoot       = "#AAAA00"
	colorNewRoot    = "#FFFF00"
)

var (
	PendingEpoch = big.NewInt(-1)
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

	cfg.OnlyEpoch = (mode == "epoch")

	captureDagEvents(cfg)
}

func captureDagEvents(cfg Config) {

	url := fmt.Sprintf("http://%s:%d/", cfg.RPCHost, cfg.RPCPort)
	conn, err := rpc.Dial(url)
	if err != nil {
		log.Panicf("Can not connect RPC: %s\n", err)
	}
	rpc := ftmclient.NewClient(conn)
	ctx := context.TODO()

	consensusCfg := integration.Configs{
		Opera:       gossip.DefaultConfig(cachescale.Identity),
		Lachesis:    abft.DefaultConfig(),
		VectorClock: vecmt.DefaultConfig(cachescale.Identity),
	}

	var (
		prevEpoch idx.Epoch
		prevGraph *dot.Graph
		prevElems *types.GraphData
	)
	for {
		epoch, graph, elems, err := readCurrentEpochDag(ctx, rpc, consensusCfg)
		if err != nil {
			log.Printf("Error while read DAG: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if prevEpoch == epoch {
			// compare graphs elements and mark red changes
			elems.MarkChanges(prevElems, "red", "2.5", colorRoot, colorNewRoot, colorAtropos, colorNewAtropos)
		} else {
			log.Println("New epoch out")
		}

		if !cfg.OnlyEpoch || prevEpoch != epoch {
			flushToFile(&cfg, prevEpoch, prevGraph)
		}

		prevEpoch = epoch
		prevGraph = graph
		prevElems = elems

		log.Printf("Capture loop done, got %d nodes\n", elems.NodesCount())
		time.Sleep(5 * time.Second)
	}
}
