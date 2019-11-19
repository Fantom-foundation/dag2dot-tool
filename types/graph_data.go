package types

import "git.sfxdx.ru/fantom/dot-tool/dot"

type GraphData struct {
	nodes map[string]*dot.Node
	edges map[string]*dot.Edge
}

func (gd *GraphData) AddNode(n *dot.Node) {
	if gd.nodes == nil {
		gd.nodes = make(map[string]*dot.Node)
	}

	gd.nodes[n.Name()] = n
}

func (gd *GraphData) AddEdge(e *dot.Edge) {
	if gd.edges == nil {
		gd.edges = make(map[string]*dot.Edge)
	}

	key := e.Source().Name()+"->"+e.Destination().Name()
	gd.edges[key] = e
}

func (gd *GraphData) MarkChanges(old *GraphData, newColor, newPenWidth, colorRoot, colorNewRoot, colorOldRoot string) {
	if old == nil {
		return
	}

	for k, n := range gd.nodes {
		oldNode, ok := old.nodes[k]
		if !ok {
			n.Set("color", newColor)
			n.Set("penwidth", newPenWidth)
		} else {
			if n.Get("fillcolor") != oldNode.Get("fillcolor") {
				n.Set("style", "filled")
				if n.Get("fillcolor") == colorRoot {
					n.Set("fillcolor", colorNewRoot)
				} else {
					n.Set("fillcolor", colorOldRoot)
				}
			}
		}
	}

	for k, e := range gd.edges {
		_, ok := old.edges[k]
		if !ok {
			e.Set("color", newColor)
			e.Set("penwidth", newPenWidth)
		}
	}
}
