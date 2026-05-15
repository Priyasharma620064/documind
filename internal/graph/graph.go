package graph

import (
	"sync"

	"github.com/priya-sharma/documind/internal/models"
)

// Graph represents the semantic knowledge graph.
type Graph struct {
	mu    sync.RWMutex
	Nodes map[string]*models.GraphNode
	Edges []models.GraphEdge
}

// NewGraph creates a new empty Graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*models.GraphNode),
		Edges: []models.GraphEdge{},
	}
}

// AddNode adds a node to the graph if it doesn't exist.
func (g *Graph) AddNode(node *models.GraphNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Nodes[node.ID] = node
}

// AddEdge adds a directed edge between two nodes.
func (g *Graph) AddEdge(sourceID, targetID string, edgeType models.EdgeType, weight float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Edges = append(g.Edges, models.GraphEdge{
		SourceID: sourceID,
		TargetID: targetID,
		Type:     edgeType,
		Weight:   weight,
	})
}

// GetRelatedNodes finds all nodes directly connected to the given node ID.
func (g *Graph) GetRelatedNodes(id string) []*models.GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var related []*models.GraphNode
	for _, edge := range g.Edges {
		if edge.SourceID == id {
			if node, ok := g.Nodes[edge.TargetID]; ok {
				related = append(related, node)
			}
		} else if edge.TargetID == id {
			if node, ok := g.Nodes[edge.SourceID]; ok {
				related = append(related, node)
			}
		}
	}
	return related
}
