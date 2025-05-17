package sqlite

import (
	"testing"

	"github.com/sunary/sqlize"
)

// TestParserSingleTable tests that Sqlize can parse a sqlite schema with one table.
func TestParserSingleTable(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSql := readFile(t, schemaWithOneTable)
	if err := sqlizeCurrent.FromString(schemaSql); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
}

// TestParserMultipleTables tests that Sqlize can parse a sqlite schema with foreign keys.
func TestParserMultipleTables(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSql := readFile(t, schemaWithTwoTables)
	if err := sqlizeCurrent.FromString(schemaSql); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
}
