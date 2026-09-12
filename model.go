package main

import (
	"sort"

	"github.com/halfcrazy/ovsdbviz/ovsdb"
)

// Graph 是面向 JSON/HTML 输出的图模型：表为节点，uuid 引用为边。
// 列的 Type 原样透传 RFC 7047 的类型描述（string 或 map），由前端渲染。
type Graph struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

type Node struct {
	ID      string   `json:"id"`
	Root    bool     `json:"root,omitempty"`
	Columns []Column `json:"columns"`
}

type Column struct {
	Name  string      `json:"name"`
	Type  interface{} `json:"type"`
	Index bool        `json:"index,omitempty"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`          // 引用所在列名
	Kind   string `json:"kind"`           // key 或 value，对应 RFC 7047 引用位置
	Weak   bool   `json:"weak,omitempty"` // refType == weak 的弱引用
}

// buildGraph 从 schema 提取节点/边，输出按表名、列名排序保证确定性。
func buildGraph(schema *ovsdb.DatabaseSchema) Graph {
	g := Graph{Name: schema.Name, Version: schema.Version}

	tableNames := make([]string, 0, len(schema.Tables))
	for name := range schema.Tables {
		tableNames = append(tableNames, name)
	}
	sort.Strings(tableNames)

	for _, tableName := range tableNames {
		table := schema.Tables[tableName]
		node := Node{ID: tableName, Root: table.IsRoot}

		columnNames := make([]string, 0, len(table.Columns))
		for name := range table.Columns {
			columnNames = append(columnNames, name)
		}
		sort.Strings(columnNames)

		for _, columnName := range columnNames {
			column := table.Columns[columnName]
			node.Columns = append(node.Columns, Column{
				Name:  columnName,
				Type:  column.Type,
				Index: table.IsIndex(columnName),
			})

			for kind, refTable := range column.RefersTo() {
				g.Edges = append(g.Edges, Edge{
					Source: tableName,
					Target: refTable,
					Label:  columnName,
					Kind:   kind,
					Weak:   isWeakRef(column.Type, kind),
				})
			}
		}
		g.Nodes = append(g.Nodes, node)
	}

	return g
}

// isWeakRef 判断引用是否为 RFC 7047 的 weak reference。
func isWeakRef(columnType interface{}, kind string) bool {
	typeMap, ok := columnType.(map[string]interface{})
	if !ok {
		return false
	}
	part, ok := typeMap[kind].(map[string]interface{})
	if !ok {
		return false
	}
	return part["refType"] == "weak"
}
