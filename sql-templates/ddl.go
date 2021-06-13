package sql_templates

import (
	"strings"
)

type Sql struct {
	IsPostgres bool
	IsLower    bool
}

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

func (s Sql) CreateTableStm() string {
	return s.apply("CREATE TABLE %s (\n%s\n);")
}

func (s Sql) DropTableStm() string {
	return s.apply("DROP TABLE IF EXISTS %s;")
}

func (s Sql) RenameTableStm() string {
	return s.apply("ALTER TABLE %s RENAME TO %s;")
}

func (s Sql) AlterTableAddColumnFirstStm() string {
	return s.apply("ALTER TABLE %s ADD COLUMN %s FIRST;")
}

func (s Sql) AlterTableAddColumnAfterStm() string {
	return s.apply("ALTER TABLE %s ADD COLUMN %s AFTER %s;")
}

func (s Sql) AlterTableDropColumnStm() string {
	return s.apply("ALTER TABLE %s DROP COLUMN %s;")
}

func (s Sql) AlterTableModifyColumnStm() string {
	return s.apply("ALTER TABLE %s MODIFY COLUMN %s;")
}

func (s Sql) AlterTableRenameColumnStm() string {
	return s.apply("ALTER TABLE %s RENAME COLUMN %s TO %s;")
}

func (s Sql) CreatePrimaryKeyStm() string {
	return s.apply("ALTER TABLE %s ADD PRIMARY KEY(%s);")
}

func (s Sql) CreateIndexStm(indexType string) string {
	if indexType != "" {
		return s.apply("CREATE INDEX %s ON %s(%s) USING " + strings.ToUpper(indexType) + ";")
	}

	return s.apply("CREATE INDEX %s ON %s(%s);")
}

func (s Sql) CreateUniqueIndexStm(indexType string) string {
	if indexType != "" {
		return s.apply("CREATE UNIQUE INDEX %s ON %s(%s) USING " + strings.ToUpper(indexType) + ";")
	}

	return s.apply("CREATE UNIQUE INDEX %s ON %s(%s);")
}

func (s Sql) DropPrimaryKeyStm() string {
	return s.apply("ALTER TABLE %s DROP PRIMARY KEY;")
}

func (s Sql) DropIndexStm() string {
	return s.apply("DROP INDEX %s ON %s;")
}

func (s Sql) AlterTableRenameIndexStm() string {
	return s.apply("ALTER TABLE %s RENAME INDEX %s TO %s;")
}
