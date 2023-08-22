package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/dag2dot-tool/dot"
)

func flushToFile(cfg *Config, epoch idx.Epoch, g *dot.Graph) {
	prefix := g.Name()
	if cfg.OnlyEpoch {
		prefix = fmt.Sprintf("DAG-EPOCH-%d", epoch)
	}
	fileBase := filepath.Join(cfg.OutPath, prefix)

	// save *.dot
	fileDot := fileBase + ".dot"
	fl, err := os.Create(fileDot)
	if err != nil {
		log.Panicf("Can not create file '%s': %s\n", fileDot, err)
	}
	_, err = fl.WriteString(g.String())
	if err != nil {
		log.Panicf("Can not write data to file '%s': %s\n", fileDot, err)
	}
	_ = fl.Close()

	// render *.png
	if !cfg.RenderFile {
		return
	}
	filePng := fileBase + ".png"
	_, err = exec.Command("dot", "-Tpng", fileDot, "-o", filePng).Output()
	if err != nil {
		log.Panicf("Can not write img to file '%s': %s\n", filePng, err)
	}
}
