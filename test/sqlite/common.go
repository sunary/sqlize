package sqlite

import (
	"os"
	"strings"
	"testing"
)

const (
	schemaWithOneTable  = "./testdata/schema_one_table.sql"
	schemaWithTwoTables = "./testdata/schema_two_tables.sql"
)

func assertContains(t *testing.T, str, substr, message string) {
	if !strings.Contains(str, substr) {
		t.Errorf("%s: expected to find '%s' in:\n%s", message, substr, str)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(data)
}
