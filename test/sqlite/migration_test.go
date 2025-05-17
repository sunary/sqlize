package sqlite

import (
	"os"
	"testing"

	"github.com/sunary/sqlize"
)

// TestSqliteParser tests that Sqlize can generate a migration script for the simplest schema.
func TestMigrationGeneratorSingleTable(t *testing.T) {
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

	runVariousMigrationFunctions(t, sqlizeCurrent)
}

func runVariousMigrationFunctions(t *testing.T, s *sqlize.Sqlize) {
	s.StringUp()
	s.StringDown()

	s.StringUpWithVersion(0, false)
	s.StringDownWithVersion(0)

	s.StringUpWithVersion(123, false)
	s.StringDownWithVersion(123)
}
