package sqlite

import (
	"fmt"
	"testing"

	"github.com/sunary/sqlize"
)

// TestSqliteParser tests that Sqlize can generate a migration script for the simplest schema.
func TestMigrationGeneratorSingleTable(t *testing.T) {
	sqlizeCurrent := sqlize.NewSqlize(
		sqlize.WithSqlite(),
	)

	schemaSql := readFile(t, schemaWithOneTable)
	if err := sqlizeCurrent.FromString(schemaSql); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	runVariousMigrationFunctions(t, sqlizeCurrent)
}

func runVariousMigrationFunctions(t *testing.T, s *sqlize.Sqlize) {
	upSQL := s.StringUp()
	downSQL := s.StringDown()

	// Validate generated migration scripts
	assertContains(t, upSQL, "CREATE TABLE", "Up migration should create the table")
	assertContains(t, upSQL, "AUTOINCREMENT", "Up migration should include AUTOINCREMENT")
	assertContains(t, upSQL, "CHECK (\"age\" >= 18)", "Up migration should include CHECK constraint")
	assertContains(t, upSQL, "UNIQUE", "Up migration should include UNIQUE values")
	assertContains(t, upSQL, "DEFAULT", "Up migration should include DEFAULT values")

	assertContains(t, downSQL, "DROP TABLE", "Down migration should drop the table")

	upWithVersionSQL := s.StringUpWithVersion(0, false)
	downWithVersionSQL := s.StringDownWithVersion(0)

	// Validate versioned migration scripts
	assertContains(t, upWithVersionSQL, "CREATE TABLE IF NOT EXISTS schema_migrations", "Initial migration should create the migrations table")
	assertContains(t, downWithVersionSQL, "DROP TABLE IF EXISTS schema_migrations", "Down migration from before initial should drop the migrations table")

	version := s.HashValue()
	upWithVersionNumberSQL := s.StringUpWithVersion(version, false)

	// Validate versioned migration scripts
	expectedVersionInsert := fmt.Sprintf("INSERT INTO schema_migrations (version, dirty) VALUES (%d, false);", version)

	assertContains(t, upWithVersionNumberSQL, expectedVersionInsert, "Versioned up migration should include version comment")
}
