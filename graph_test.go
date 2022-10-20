package graph

import (
	"testing"
)

func TestGraph_AddEdge(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	b := Vertex[string]{"B"}
	g.AddVertex(&a)
	g.AddVertex(&b)
	tag := "tag"
	pTag := &tag

	err := g.AddEdge(&a, &b, 1, pTag)
	if err != nil {
		t.Error(err)
	}
}

func TestGraph_ContainsNode(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	g.AddVertex(&a)

	result := g.ContainsVertex(&a)
	if !result {
		t.Error("could not find vertex in graph")
	}
}

func TestGraph_ContainsEdge(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	b := Vertex[string]{"B"}
	g.AddVertex(&a)
	g.AddVertex(&b)
	tag := "tag"
	pTag := &tag

	err := g.AddEdge(&a, &b, 1, pTag)
	if err != nil {
		t.Error(err)
	}

	result := g.containsEdge(&a, &b, pTag)
	if !result {
		t.Error("could not find edge in graph")
	}
}

func TestGraph_ContainsSymmetricEdge(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	b := Vertex[string]{"B"}
	g.AddVertex(&a)
	g.AddVertex(&b)
	tag := "tag"
	pTag := &tag

	// no edges
	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	// A->B
	_ = g.AddEdge(&a, &b, 1, pTag)
	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	g.RemoveEdge(&a, &b, pTag)

	// B->A
	_ = g.AddEdge(&b, &a, 1, pTag)
	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); contains {
		t.Error("should not contain A<->B")
	}

	g.RemoveEdge(&b, &a, pTag)

	// A<->B
	if err := g.AddSymmetricEdge(&a, &b, 5, pTag); err != nil {
		t.Error("unable to add symmetric edge A<->B")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); !contains {
		t.Error("should contain A<->B, but does not!")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); !contains {
		t.Error("should contain A<->B, but does not!")
	}

	g.RemoveSymmetricEdge(&a, &b, pTag)

	// B<->A
	if err := g.AddSymmetricEdge(&b, &a, 5, pTag); err != nil {
		t.Error("unable to add symmetric edge A<->B")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); !contains {
		t.Error("should contain A<->B, but does not!")
	}

	if contains := g.ContainsSymmetricEdge(&a, &b, pTag); !contains {
		t.Error("should contain A<->B, but does not!")
	}

}

func TestGraph_RemoveSymmetricEdge(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	b := Vertex[string]{"B"}
	g.AddVertex(&a)
	g.AddVertex(&b)
	tag := "tag"
	pTag := &tag

	err := g.AddSymmetricEdge(&a, &b, 5, pTag)
	if err != nil {
		t.Error("unable to add symmetric edge")
		t.Fail()
	}

	g.RemoveSymmetricEdge(&a, &b, pTag)

	if g.ContainsEdge(&a, &b, pTag) {
		t.Error("graph still contains symmetric edge A -> B")
		t.Fail()
	}
}

func TestGraph_AddSymmetricEdge(t *testing.T) {
	g := Graph[string]{}
	a := Vertex[string]{"A"}
	b := Vertex[string]{"B"}
	g.AddVertex(&a)
	g.AddVertex(&b)
	tag := "tag"
	pTag := &tag

	err := g.AddSymmetricEdge(&a, &b, 5, pTag)
	if err != nil {
		t.Error("unable to add symmetric edge")
		t.Fail()
	}

	if !g.ContainsSymmetricEdge(&a, &b, pTag) {
		t.Error("graph does not contains symmetric edge A -> B")
		t.Fail()
	}
}

func TestGraph_ShortestPath(t *testing.T) {
	g := Graph[string]{}
	a := &(Vertex[string]{"A"})
	b := &(Vertex[string]{"B"})
	c := &(Vertex[string]{"C"})
	d := &(Vertex[string]{"D"})
	g.AddVertex(a)
	g.AddVertex(b)
	g.AddVertex(c)
	g.AddVertex(d)

	_ = g.AddEdge(a, b, 1, nil)
	_ = g.AddEdge(b, c, 10, nil)
	_ = g.addEdge(a, d, 5, nil)
	_ = g.addEdge(d, c, 5, nil)

	path, err := g.ShortestPath(a, c)
	if err != nil {
		t.Error(err)
	}

	if path[0].Source.String() != "A" || path[0].Destination.String() != "D" {
		t.Error("Path does not contain A -> D")
	}

	if path[1].Source.String() != "D" || path[1].Destination.String() != "C" {
		t.Error("Path does not contain D -> C")
	}
}
