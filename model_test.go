package main

import (
	"testing"

	"github.com/halfcrazy/ovsdbviz/ovsdb"
)

func TestBuildGraphVswitch(t *testing.T) {
	schema, err := ovsdb.NewDatabaseSchema(ovsdb.SchemaOption{SchemaPath: "examples/vswitch.ovsschema"})
	if err != nil {
		t.Fatal(err)
	}
	g := buildGraph(schema)

	if g.Name != "Open_vSwitch" {
		t.Fatalf("name = %q", g.Name)
	}
	if len(g.Nodes) != 19 {
		t.Fatalf("nodes = %d, want 19", len(g.Nodes))
	}
	if len(g.Edges) != 22 {
		t.Fatalf("edges = %d, want 22", len(g.Edges))
	}

	find := func(src, dst, label string) *Edge {
		for i, e := range g.Edges {
			if e.Source == src && e.Target == dst && e.Label == label {
				return &g.Edges[i]
			}
		}
		return nil
	}

	// Bridge.mirrors 是 strong reference
	if e := find("Bridge", "Mirror", "mirrors"); e == nil || e.Weak || e.Kind != "key" {
		t.Fatalf("Bridge->Mirror mirrors edge wrong: %+v", e)
	}
	// Mirror.output_port 是 weak reference（RFC 7047 refType）
	if e := find("Mirror", "Port", "output_port"); e == nil || !e.Weak {
		t.Fatalf("Mirror->Port output_port should be weak: %+v", e)
	}
	// 排序后首个节点应为 AutoAttach
	if g.Nodes[0].ID != "AutoAttach" {
		t.Fatalf("nodes not sorted, first = %q", g.Nodes[0].ID)
	}
}
