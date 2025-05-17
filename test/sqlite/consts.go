package sqlite

import (
	"strings"
	"testing"
)

const (
	schemaOneTable  = "data/schema_one_table.sql"
	schemaTwoTables = "data/schema_two_tables.sql"
)

func assertContains(t *testing.T, str, substr, message string) {
	if !strings.Contains(str, substr) {
		t.Errorf("%s: expected to find '%s' in:\n%s", message, substr, str)
	}
}
