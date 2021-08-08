package sql_templates

import (
	"strings"
)

// Sql ...
type Sql struct {
	IsPostgres bool
	IsLower    bool
}

// NewSql ...
func NewSql(isPostgres, isLower bool) *Sql {
	return &Sql{
		IsPostgres: isPostgres,
		IsLower:    isLower,
	}
}

func (s Sql) apply(t string) string {
	if s.IsLower {
		return strings.ToLower(t)
	}

	return t
}

// CreateTableStm ...
func (s Sql) CreateTableStm() string {
	return s.apply("CREATE TABLE %s (\n%s\n);")
}

// DropTableStm ...
func (s Sql) DropTableStm() string {
	return s.apply("DROP TABLE IF EXISTS %s;")
}

// RenameTableStm ...
func (s Sql) RenameTableStm() string {
	return s.apply("ALTER TABLE %s RENAME TO %s;")
}

// AlterTableAddColumnFirstStm ...
func (s Sql) AlterTableAddColumnFirstStm() string {
	return s.apply("ALTER TABLE %s ADD COLUMN %s FIRST;")
}

// AlterTableAddColumnAfterStm ...
func (s Sql) AlterTableAddColumnAfterStm() string {
	return s.apply("ALTER TABLE %s ADD COLUMN %s AFTER %s;")
}

// AlterTableDropColumnStm ...
func (s Sql) AlterTableDropColumnStm() string {
	return s.apply("ALTER TABLE %s DROP COLUMN %s;")
}

// AlterTableModifyColumnStm ...
func (s Sql) AlterTableModifyColumnStm() string {
	return s.apply("ALTER TABLE %s MODIFY COLUMN %s;")
}

// AlterTableRenameColumnStm ...
func (s Sql) AlterTableRenameColumnStm() string {
	return s.apply("ALTER TABLE %s RENAME COLUMN %s TO %s;")
}

// CreatePrimaryKeyStm ...
func (s Sql) CreatePrimaryKeyStm() string {
	return s.apply("ALTER TABLE %s ADD PRIMARY KEY(%s);")
}

// CreateIndexStm ...
func (s Sql) CreateIndexStm(indexType string) string {
	if indexType != "" {
		return s.apply("CREATE INDEX %s ON %s(%s) USING " + strings.ToUpper(indexType) + ";")
	}

	return s.apply("CREATE INDEX %s ON %s(%s);")
}

// CreateUniqueIndexStm ...
func (s Sql) CreateUniqueIndexStm(indexType string) string {
	if indexType != "" {
		return s.apply("CREATE UNIQUE INDEX %s ON %s(%s) USING " + strings.ToUpper(indexType) + ";")
	}

	return s.apply("CREATE UNIQUE INDEX %s ON %s(%s);")
}

// DropPrimaryKeyStm ...
func (s Sql) DropPrimaryKeyStm() string {
	return s.apply("ALTER TABLE %s DROP PRIMARY KEY;")
}

// DropIndexStm ...
func (s Sql) DropIndexStm() string {
	return s.apply("DROP INDEX %s ON %s;")
}

// AlterTableRenameIndexStm ...
func (s Sql) AlterTableRenameIndexStm() string {
	return s.apply("ALTER TABLE %s RENAME INDEX %s TO %s;")
}
