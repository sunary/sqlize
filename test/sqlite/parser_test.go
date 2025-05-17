package sqlite

import (
	"os"
	"testing"

	"github.com/sunary/sqlize"
)

const (
	schemaOneTable  = "data/schema_one_table.sql"
	schemaTwoTables = "data/schema_two_tables.sql"
)

// TestSqliteParser tests that Sqlize can parse a sqlite schema with one table.
func TestParserSingleTable(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSqlBytes, err := os.ReadFile(schemaOneTable)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}
	if err := sqlizeCurrent.FromString(string(schemaSqlBytes)); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
}

// TestSqliteParser tests that Sqlize can parse a sqlite schema with foreign keys.
func TestParserMultipleTables(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSqlBytes, err := os.ReadFile(schemaTwoTables)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}
	if err := sqlizeCurrent.FromString(string(schemaSqlBytes)); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
}
