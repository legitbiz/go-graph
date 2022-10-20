package graph

import (
	"container/heap"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"math"
	"sync"
)

type Vertex[TValue comparable] struct {
	value TValue
}

func (n *Vertex[TValue]) String() string {
	return fmt.Sprintf("%v", n.value)
}

type weightedEdge[TValue comparable] struct {
	destination *Vertex[TValue]
	weight      uint
	tag         *string
}

func (we weightedEdge[TValue]) Destination() TValue {
	return we.destination.value
}

func (we weightedEdge[TValue]) Weight() uint {
	return we.weight
}

func (we weightedEdge[TValue]) Tag() *string {
	return we.tag
}

type WeightedEdge[TValue comparable] interface {
	Destination() TValue
	Weight() uint
	Tag() *string
}

type PathEdge[TValue comparable] struct {
	Source      *Vertex[TValue]
	Destination *Vertex[TValue]
	Weight      uint
	Tag         *string
}

func (pe PathEdge[TValue]) String() string {
	if pe.Tag == nil {
		return fmt.Sprintf("%s -> %s, Cost: %d",
			pe.Source.String(),
			pe.Destination.String(),
			pe.Weight)
	}
	return fmt.Sprintf("%s -> %s, Cost: %d, tag: '%s'",
		pe.Source.String(),
		pe.Destination.String(),
		pe.Weight,
		*pe.Tag)
}

// Graph - a directed, weighted graph.
type Graph[TValue comparable] struct {
	vertices []*Vertex[TValue]
	edges    map[Vertex[TValue]][]weightedEdge[TValue]
	lock     sync.RWMutex
}

// AddVertex adds a vertex to the graph without any edges. If the vertex already
// exists, no action is taken.
func (g *Graph[TValue]) AddVertex(v *Vertex[TValue]) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if !slices.Contains(g.vertices, v) {
		g.vertices = append(g.vertices, v)
	}
}

// RemoveSymmetricEdge removes src->dest and dest->src
func (g *Graph[TValue]) RemoveSymmetricEdge(src, dest *Vertex[TValue], tag *string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.removeEdge(src, dest, tag)
	g.removeEdge(dest, src, tag)
}

// RemoveEdge removes only the edge src->dest. It will not remove dest->src
func (g *Graph[TValue]) RemoveEdge(src, dest *Vertex[TValue], tag *string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.removeEdge(src, dest, tag)
}

func (g *Graph[TValue]) removeEdge(src, dest *Vertex[TValue], tag *string) {
	if g.vertices == nil {
		return
	}

	if !g.containsEdge(src, dest, tag) {
		return
	}

	f := func(we weightedEdge[TValue]) bool {
		return we.destination == dest
	}

	if idx := slices.IndexFunc(g.edges[*src], f); idx >= 0 {
		g.edges[*src] = slices.Delete(g.edges[*src], idx, idx+1)
	}
}

// AddSymmetricEdge creates an edge from src->dest and dest->src with the same weight
// for both directions
func (g *Graph[TValue]) AddSymmetricEdge(src, dest *Vertex[TValue], weight uint, tag *string) error {
	if err := g.isEdgeValid(src, dest, weight); err != nil {
		return err
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if err := g.addEdge(src, dest, weight, tag); err != nil {
		return err
	}

	if err := g.addEdge(dest, src, weight, tag); err != nil {
		// rollback the previous add then return err
		g.removeEdge(src, dest, tag)
		return err
	}

	return nil
}

// AddEdge creates a directed edge from src->dest with a non-zero weight and
// an optional tag. Supply `nil` if there's no tag.
func (g *Graph[TValue]) AddEdge(src *Vertex[TValue], dest *Vertex[TValue], weight uint, tag *string) error {
	if err := g.isEdgeValid(src, dest, weight); err != nil {
		return err
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	return g.addEdge(src, dest, weight, tag)
}

func (g *Graph[TValue]) isEdgeValid(src, dest *Vertex[TValue], weight uint) error {
	if weight == 0 {
		return errors.New("weight cannot be 0")
	}

	if src == nil {
		return errors.New("src cannot be nil")
	}

	if dest == nil {
		return errors.New("dest cannot be nil")
	}

	if !g.containsVertex(src) {
		return errors.New("graph does not contain src")
	}

	if !g.containsVertex(dest) {
		return errors.New("graph does not contain dest")
	}

	return nil
}

// addEdge is the helper method for both AddEdge and AddSymmetricEdge.
func (g *Graph[TValue]) addEdge(src *Vertex[TValue], dest *Vertex[TValue], weight uint, tag *string) error {
	if !slices.Contains(g.vertices, src) {
		return errors.New("unable to locate src in graph")
	}

	if !slices.Contains(g.vertices, dest) {
		return errors.New("unable to locate dest in graph")
	}

	if g.edges == nil {
		g.edges = make(map[Vertex[TValue]][]weightedEdge[TValue])
	}

	// check if src's edges contains dest
	if g.containsEdge(src, dest, tag) {
		return errors.New("this edge is already present")
	}

	// otherwise add src->dest
	g.edges[*src] = append(g.edges[*src], weightedEdge[TValue]{dest, weight, tag})

	return nil
}

// ContainsVertex checks if the graph contains a vertex
func (g *Graph[TValue]) ContainsVertex(v *Vertex[TValue]) bool {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.containsVertex(v)
}

func (g *Graph[TValue]) containsVertex(v *Vertex[TValue]) bool {
	return slices.Contains(g.vertices, v)
}

// ContainsEdge checks if the graph contains the edge src->dest
func (g *Graph[TValue]) ContainsEdge(src, dest *Vertex[TValue], tag *string) bool {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.containsEdge(src, dest, tag)
}

// GetEdge retrieves an edge from the graph.
func (g *Graph[TValue]) GetEdge(src, dest *Vertex[TValue], tag *string) (WeightedEdge[TValue], error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.getEdge(src, dest, tag)
}

func (g *Graph[TValue]) getEdge(src, dest *Vertex[TValue], tag *string) (WeightedEdge[TValue], error) {
	if !g.ContainsEdge(src, dest, tag) {
		return nil, errors.New("unable to find edge")
	}

	es := g.edges[*src]
	for _, we := range es {
		if we.destination == dest && *we.tag == *tag {
			return we, nil
		}
	}

	return nil, errors.New("unable to find edge")
}

// ContainsSymmetricEdge checks if the graph contains an edge in both
// directions AND if that edge has the same weight in both directions.
func (g *Graph[TValue]) ContainsSymmetricEdge(src, dest *Vertex[TValue], tag *string) bool {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if !g.containsEdge(src, dest, tag) {
		return false
	}

	if !g.containsEdge(dest, src, tag) {
		return false
	}

	e1, err := g.getEdge(src, dest, tag)
	if err != nil {
		return false
	}

	e2, err := g.getEdge(dest, src, tag)
	if err != nil {
		return false
	}

	return e1.Weight() == e2.Weight()
}

func (g *Graph[TValue]) containsEdge(src, dest *Vertex[TValue], tag *string) bool {
	edges, exists := g.edges[*src]
	if !exists {
		return false
	}

	for _, edge := range edges {
		if edge.destination != dest {
			continue
		}
		// found the destination, now check tags

		// both tags are nil that's the same edge, we're done!
		if edge.tag == nil && tag == nil {
			return true
		}

		// tag isn't nil, edge isn't nil.
		// We know that edge.destination does equal destination.
		// All that's left is to check the tags!
		if *edge.tag == *tag {
			return true
		}
	}

	return false
}

// ShortestPath is an implementation of Dijkstra's algorithm for a single
// src->dest route.
func (g *Graph[TValue]) ShortestPath(src, dest *Vertex[TValue]) ([]PathEdge[TValue], error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.shortestPath(src, dest)
}

type queueItem[TValue comparable] struct {
	source *Vertex[TValue]
	tag    *string
	weight uint
}

// shortestPath is an implementation of Dijkstra's algorithm
//
// Wikipedia claims:
//
//	1  function Dijkstra(Graph, source, target):
//	2
//	3      for each vertex v in Graph.Vertices:
//	4          dist[v] ← INFINITY
//	5          prev[v] ← UNDEFINED
//	6          add v to Q
//	7      dist[source] ← 0
//	8
//	9      while Q is not empty:
//
// 10          u ← vertex in Q with min dist[u]
//
//	if u = target
//	   break
//
// 11          remove u from Q
// 12
// 13          for each neighbor v of u still in Q:
// 14              alt ← dist[u] + Graph.Edges(u, v)
// 15              if alt < dist[v]:
// 16                  dist[v] ← alt
// 17                  prev[v] ← u
// 18
//
//	1  S ← empty sequence
//	2  u ← target
//	3  if prev[u] is defined or u = source:          // Do something only if the vertex is reachable
//	4      while u is defined:                       // Construct the shortest path with a stack S
//	5          insert u at the beginning of S        // Push the vertex onto the stack
//	6          u ← prev[u]                           // Traverse from target to source
func (g *Graph[TValue]) shortestPath(src, dest *Vertex[TValue]) ([]PathEdge[TValue], error) {

	// Set the distance to src to 0
	distance := make(map[*Vertex[TValue]]uint)
	distance[src] = 0

	// create a vertex priority queue
	q := &vertexDistanceHeap[TValue]{}

	// for each vertex v in Graph.Vertices:
	for _, v := range g.vertices {
		if *v != *src {
			// dist[v] ← INFINITY
			distance[v] = math.MaxInt
			// skipping prev[v] ← UNDEFINED
		}

		// Q.add_with_priority(v, dist[v])
		heap.Push(q, vertexDistance[TValue]{vertex: v, distance: distance[v]})
	}

	prev := make(map[*Vertex[TValue]]queueItem[TValue])

	// while Q is not empty:
	for q.Len() != 0 {
		// u ← vertex in Q with min dist[u]
		u := heap.Pop(q).(vertexDistance[TValue])

		//
		if u.vertex == dest {
			break
		}

		neighbors := g.edges[*u.vertex]
		// for each neighbor v of u
		for _, uToV := range neighbors {
			v := uToV.destination
			// alt ← dist[u] + Graph.Edges(u, v)
			alt := distance[u.vertex] + uToV.weight
			if distance[v] > alt {
				// dist[v] ← alt
				distance[v] = alt
				// prev[v] ← u
				prev[v] = queueItem[TValue]{u.vertex, uToV.tag, uToV.weight}
				// Q.decrease_priority(v, alt)
				q.updateDistance(v, alt)
			}
		}
		heap.Init(q)
	}

	// And now we build up the shortest path!

	// S ← empty sequence
	path := []PathEdge[TValue]{}
	// u ← target
	u := dest

	for {
		// if prev[u] is defined or u = source:
		qn, ok := prev[u]
		if !ok {
			break
		}

		// insert u at the beginning of S
		t := make([]PathEdge[TValue], len(path)+1)
		t[0] = PathEdge[TValue]{qn.source, u, qn.weight, qn.tag}
		copy(t[1:], path)
		path = t

		// u ← prev[u]
		u = prev[u].source
	}

	return path, nil
}

// vertexDistance implements a min-heap for calculating the shortest-path between
// two vertices in a graph
type vertexDistance[T comparable] struct {
	vertex   *Vertex[T]
	distance uint
}

type vertexDistanceHeap[T comparable] []vertexDistance[T]

func (h *vertexDistanceHeap[T]) Len() int {
	return len(*h)
}

func (h *vertexDistanceHeap[T]) Less(i, j int) bool {
	return (*h)[i].distance < (*h)[j].distance
}

func (h *vertexDistanceHeap[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *vertexDistanceHeap[T]) Push(x interface{}) {
	*h = append(*h, x.(vertexDistance[T]))
}

func (h *vertexDistanceHeap[T]) Pop() interface{} {
	heapSize := len(*h)
	lastVertex := (*h)[heapSize-1]
	*h = (*h)[0 : heapSize-1]
	return lastVertex
}

func (h *vertexDistanceHeap[T]) updateDistance(id *Vertex[T], val uint) {
	for i := 0; i < len(*h); i++ {
		if (*h)[i].vertex == id {
			(*h)[i].distance = val
			break
		}
	}
}
