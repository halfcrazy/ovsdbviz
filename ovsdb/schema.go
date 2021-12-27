package ovsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"io/ioutil"
	"net"
	"os"
)

type DatabaseSchema struct {
	Name     string                 `json:"name"`
	Version  string                 `json:"version"`
	Checksum string                 `json:"cksum,omitempty"`
	Tables   map[string]TableSchema `json:"tables"`
}

// OrderedColumns returns all column names ordered by a numeric index for each
// table in the schema
func (database DatabaseSchema) OrderedColumns() map[string][]string {
	tableColumnOrder := make(map[string][]string)

	for tableName, table := range database.Tables {
		var columnOrder []string
		columnOrder = append(columnOrder, tableName)
		for columnName := range table.Columns {
			columnOrder = append(columnOrder, columnName)
		}
		tableColumnOrder[tableName] = columnOrder
	}

	return tableColumnOrder

}

type TableSchema struct {
	Columns     map[string]ColumnSchema `json:"columns"`
	ColumnOrder []string
	MaxRows     int        `json:"maxrows,omitempty"`
	IsRoot      bool       `json:"isRoot,omitempty"`
	Indexes     [][]string `json:"indexes,omitempty"`
}

// FIXME: Doesn't support composite indexes
func (table TableSchema) IsIndex(columnName string) bool {
	for i, _ := range table.Indexes {
		for _, v := range table.Indexes[i] {
			if columnName == v {
				return true
			}
		}
	}

	return false
}

type ColumnSchema struct {
	Type interface{} `json:"type"`
}

func getRefTable(key string, typeColumn interface{}) string {
	typeColumnMap := typeColumn.(map[string]interface{})

	if valueInterface, exists := typeColumnMap[key]; exists {
		switch valueInterface.(type) {
		case map[string]interface{}:
			value := valueInterface.(map[string]interface{})
			if refTable, exists := value["refTable"]; exists {
				return refTable.(string)
			}
		}
	}

	return ""
}

func (column ColumnSchema) RefersTo() map[string]string {

	references := make(map[string]string)

	switch column.Type.(type) {
	case map[string]interface{}:
		keyRefTable := getRefTable("key", column.Type)
		if keyRefTable != "" {
			references["key"] = keyRefTable
		}

		valueRefTable := getRefTable("value", column.Type)
		if valueRefTable != "" {
			references["value"] = valueRefTable
		}
	}

	return references
}

type SchemaOption struct {
	Address    string
	DB         string
	SchemaPath string
}

func NewDatabaseSchema(opt SchemaOption) (*DatabaseSchema, error) {
	if opt.SchemaPath != "" {
		return getSchemaFromFile(opt.SchemaPath)
	}
	return getSchemaFromRpc(opt.Address, opt.DB)
}

func getSchemaFromFile(schemaPath string) (*DatabaseSchema, error) {
	fp, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	database := DatabaseSchema{}

	if err = json.Unmarshal(fp, &database); err != nil {
		return nil, err
	}

	return &database, nil
}

var ErrNotConnected = errors.New("not connected")

func newRPCClient(address string) (*rpc2.Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	client := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(conn))
	client.SetBlocking(true)
	go client.Run()
	return client, nil
}

func getSchemaFromRpc(address, dbName string) (*DatabaseSchema, error) {
	client, err := newRPCClient(address)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	dbs, err := listDbs(client)
	if err != nil {
		return nil, err
	}
	found := false
	for _, db := range dbs {
		if dbName == db {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("db %s not found, available dbs are %v\n", dbName, dbs)
		os.Exit(1)
	}
	return getSchema(client, dbName)
}

func getSchema(client *rpc2.Client, dbName string) (*DatabaseSchema, error) {
	args := []interface{}{dbName}
	var reply DatabaseSchema
	if err := client.Call("get_schema", args, &reply); err != nil {
		if err == rpc2.ErrShutdown {
			return nil, ErrNotConnected
		}
		return nil, err
	}
	return &reply, nil
}

func listDbs(client *rpc2.Client) ([]string, error) {
	var dbs []string
	err := client.Call("list_dbs", nil, &dbs)
	if err != nil {
		if err == rpc2.ErrShutdown {
			return nil, ErrNotConnected
		}
		return nil, fmt.Errorf("listdbs failure - %v", err)
	}
	return dbs, err
}
