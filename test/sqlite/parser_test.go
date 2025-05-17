package sqlite

import (
	"os"
	"testing"

	"github.com/sunary/sqlize"
)

const (
	schemaFileLocation = "data/schema_one_table.sql"
)

// TestSqliteParser tests that Sqlize can parse a sqlite schema.
func TestSqliteParser(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSqlBytes, err := os.ReadFile(schemaFileLocation)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}
	t.Log(string(schemaSqlBytes))
	if err := sqlizeCurrent.FromString(string(schemaSqlBytes)); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
}
