# go-graph

This implements a weighted, directional graph in Go.

## Usage

```go
// Construct an empty graph storing strings inside the node.
// The node value is a generic, TValue. TValue _must_ implement comparable.
g := Graph[string]{}

// Vertices must be added before adding any edges. If you haven't added a vertex,
// the edge add will return an error.
a := &(Vertex[string]{"A"})
b := &(Vertex[string]{"B"})
c := &(Vertex[string]{"C"})
d := &(Vertex[string]{"D"})

// AddVertex will silently continue if a Vertex already exists.
g.AddVertex(a)
g.AddVertex(b)
g.AddVertex(c)
g.AddVertex(d)

// Now that we've added the vertices, we can get to adding edges.
// I'm ignoring error checks here for clarity.
_ = g.AddEdge(a, b, 1, nil)
_ = g.AddEdge(b, c, 10, nil)
_ = g.addEdge(a, d, 5, nil)
_ = g.addEdge(d, c, 5, nil)

// calculate the shortest path between "A" and "C"
path, err := g.ShortestPath(a, c)
if err != nil {
    t.Error(err)
}
```