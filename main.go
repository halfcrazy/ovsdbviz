package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/halfcrazy/ovsdbviz/graphviz"
	"github.com/halfcrazy/ovsdbviz/ovsdb"
	"github.com/jessevdk/go-flags"
	"os"
	"sort"
	"strings"
)

//go:embed viewer.html
var viewerHTML string

const (
	tableAttrRow         = `<tr><td port="f%d" border="1" bgcolor="%s">%s</td></tr>`
	tableBackgroundColor = "turquoise1"
)

func CreateLabel(table ovsdb.TableSchema, columns []string) string {
	var labels []string

	for index, columnName := range columns {
		tableBgColor := "transparent"
		label := columnName
		if index == 0 {
			tableBgColor = tableBackgroundColor
			if table.IsRoot {
				label = fmt.Sprintf("%s (root)", label)
			}
		}

		if table.IsIndex(columnName) {
			label = fmt.Sprintf("%s (index)", label)
		}

		labels = append(labels, fmt.Sprintf(tableAttrRow, index, tableBgColor, label))
	}

	return strings.Join(labels, "")
}

func GetPortIndex(columns []string, column string) int {
	portIndex := 0 // pointing to the table name by default
	for i, columnName := range columns {
		if columnName == column {
			portIndex = i
			break
		}
	}

	return portIndex
}

type RPCOptions struct {
	DBName  string `long:"db" description:"ovs db name"`
	Address string `long:"address" description:"ovs db server address, eg 192.168.1.1:6641"`
}

type LocalOptions struct {
	SchemaPaths []string `long:"schema" description:"ovs schema file path, repeatable for html format"`
}

type CliOptions struct {
	Out        string       `long:"out" description:"output file path" required:"true"`
	Format     string       `long:"format" description:"output format: dot | json | html" choice:"dot" choice:"json" choice:"html" default:"dot"`
	Local      LocalOptions `group:"local"`
	RPC        RPCOptions   `group:"rpc"`
	Positional struct {
		SchemaPaths []string `positional-arg-name:"schema" description:"ovs schema files, same as --schema"`
	} `positional-args:"yes"`
}

var cliOptions CliOptions
var parser = flags.NewParser(&cliOptions, flags.Default)

func parseOptions() {
	if _, err := parser.Parse(); err != nil {
		fmt.Println(parser.Usage)
		os.Exit(1)
	}
	cliOptions.Local.SchemaPaths = append(cliOptions.Local.SchemaPaths, cliOptions.Positional.SchemaPaths...)
	useLocal := len(cliOptions.Local.SchemaPaths) > 0
	useRPC := cliOptions.RPC.DBName != "" || cliOptions.RPC.Address != ""
	if useLocal && useRPC {
		fmt.Println("cannot specify local and rpc in the same time")
		os.Exit(1)
	}
	if !useLocal && !useRPC {
		fmt.Println("you must specify local or rpc")
		os.Exit(1)
	}
	if useRPC && (cliOptions.RPC.DBName == "" || cliOptions.RPC.Address == "") {
		fmt.Println("rpc requires both --db and --address")
		os.Exit(1)
	}
	// rpc 每次只能取一个库的 schema，多 schema 仅 local html 支持
	if cliOptions.Format != "html" && len(cliOptions.Local.SchemaPaths) > 1 {
		fmt.Println("only html format accepts multiple --schema")
		os.Exit(1)
	}
}

// loadSchemas 加载全部输入 schema（local 可多个，rpc 仅一个）。
func loadSchemas() []*ovsdb.DatabaseSchema {
	var schemas []*ovsdb.DatabaseSchema
	if len(cliOptions.Local.SchemaPaths) > 0 {
		for _, path := range cliOptions.Local.SchemaPaths {
			schema, err := ovsdb.NewDatabaseSchema(ovsdb.SchemaOption{SchemaPath: path})
			if err != nil {
				panic(err)
			}
			schemas = append(schemas, schema)
		}
		return schemas
	}
	schema, err := ovsdb.NewDatabaseSchema(ovsdb.SchemaOption{
		Address: cliOptions.RPC.Address,
		DB:      cliOptions.RPC.DBName,
	})
	if err != nil {
		panic(err)
	}
	return []*ovsdb.DatabaseSchema{schema}
}

func main() {
	parseOptions()
	schemas := loadSchemas()

	output, err := os.Create(cliOptions.Out)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	switch cliOptions.Format {
	case "json":
		writeJSON(output, schemas[0])
	case "html":
		writeHTML(output, schemas)
	default:
		writeDot(output, schemas[0])
	}
}

func writeDot(output *os.File, schema *ovsdb.DatabaseSchema) {

	// Need to always iterate all column for a given table following the same order
	// in order to build and reference graphviz node ports
	tableColumnOrder := schema.OrderedColumns()

	graph := graphviz.NewGraph()

	// NODES
	for _, tableName := range sortedKeys(schema.Tables) {
		label := CreateLabel(schema.Tables[tableName], tableColumnOrder[tableName])
		nodeAttrs := make(map[string]string)
		nodeAttrs["shape"] = "none"
		nodeAttrs["label"] = fmt.Sprintf(`<<table border="0" cellspacing="0">%s</table>>`, label)

		graph.AddNode(tableName, nodeAttrs)
	}

	// EDGES
	for _, tableName := range sortedKeys(schema.Tables) {
		table := schema.Tables[tableName]
		for _, cn := range sortedKeys(table.Columns) {
			column := table.Columns[cn]
			references := column.RefersTo()
			if len(references) == 0 {
				continue
			}

			portIndex := GetPortIndex(tableColumnOrder[tableName], cn)

			for _, refAttribute := range sortedKeys(references) {
				reference := references[refAttribute]
				src := tableName
				srcPort := fmt.Sprintf("f%d", portIndex)
				dst := reference
				dstPort := "f0"

				edgeAttrs := make(map[string]string)
				edgeAttrs["label"] = refAttribute
				edgeAttrs["splines"] = "polyline"
				switch refAttribute {
				case "key":
					edgeAttrs["color"] = "red"
				case "value":
					edgeAttrs["color"] = "blue"
				}

				graph.AddEdge(src, srcPort, dst, dstPort, edgeAttrs)
			}
		}
	}

	_, err := output.WriteString(graph.String())
	if err != nil {
		panic(fmt.Sprintf("Error while writing output to %s: %v", cliOptions.Out, err))
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeJSON(output *os.File, schema *ovsdb.DatabaseSchema) {
	data, err := json.MarshalIndent(buildGraph(schema), "", "  ")
	if err != nil {
		panic(err)
	}
	if _, err := output.Write(data); err != nil {
		panic(fmt.Sprintf("Error while writing output to %s: %v", cliOptions.Out, err))
	}
}

// writeHTML 将全部 schema 的图数据嵌入 viewer，产出单个自包含 HTML。
func writeHTML(output *os.File, schemas []*ovsdb.DatabaseSchema) {
	graphs := make([]Graph, 0, len(schemas))
	for _, schema := range schemas {
		graphs = append(graphs, buildGraph(schema))
	}
	data, err := json.Marshal(graphs)
	if err != nil {
		panic(err)
	}
	// 防止 schema 内容意外闭合 <script>
	safe := strings.ReplaceAll(string(data), "</", `<\/`)

	page := strings.Replace(viewerHTML, "/*__DATA__*/[]", safe, 1)
	if page == viewerHTML {
		panic("viewer.html data placeholder not found")
	}
	if _, err := output.WriteString(page); err != nil {
		panic(fmt.Sprintf("Error while writing output to %s: %v", cliOptions.Out, err))
	}
}
