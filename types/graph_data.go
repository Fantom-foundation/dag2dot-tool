package types

import "github.com/Fantom-foundation/dag2dot-tool/dot"

// GraphData consists of nodes and edges
type GraphData struct {
	nodes map[string]*dot.Node
	edges map[string]*dot.Edge
}

// Add a node
func (gd *GraphData) NodesCount() int {
	return len(gd.nodes)
}

// Add a node
func (gd *GraphData) AddNode(n *dot.Node) {
	if gd.nodes == nil {
		gd.nodes = make(map[string]*dot.Node)
	}

	gd.nodes[n.Name()] = n
}

// Add an edge
func (gd *GraphData) AddEdge(e *dot.Edge) {
	if gd.edges == nil {
		gd.edges = make(map[string]*dot.Edge)
	}

	key := e.Source().Name() + "->" + e.Destination().Name()
	gd.edges[key] = e
}

// Mark the change in the graph data using new color
func (gd *GraphData) MarkChanges(old *GraphData, newColor, newPenWidth, colorRoot, colorNewRoot, colorAtropos, colorNewAtropos string) {
	if old == nil {
		return
	}

	is := func(n *dot.Node, color string) bool {
		return n.Get("fillcolor") == color
	}

	for k, newNode := range gd.nodes {
		oldNode, exists := old.nodes[k]
		if !exists {
			newNode.Set("color", newColor)
			newNode.Set("penwidth", newPenWidth)
		} else {
			if !is(oldNode, newNode.Get("fillcolor")) {
				newNode.Set("style", "filled")

				if is(newNode, colorRoot) && !is(oldNode, colorNewRoot) {
					newNode.Set("fillcolor", colorNewRoot)
				}
				if is(newNode, colorAtropos) && !is(oldNode, colorNewAtropos) {
					newNode.Set("fillcolor", colorNewAtropos)
				}
			}
		}
	}

	for k, newEdge := range gd.edges {
		_, exists := old.edges[k]
		if !exists {
			newEdge.Set("color", newColor)
			newEdge.Set("penwidth", newPenWidth)
		}
	}
}
