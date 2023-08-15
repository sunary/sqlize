package sql_templates

import (
	"strings"
)

// Sql ...
type Sql struct {
	dialect   SqlDialect
	lowercase bool
}

// NewSql ...
func NewSql(dialect SqlDialect, lowercase bool) *Sql {
	return &Sql{
		dialect:   dialect,
		lowercase: lowercase,
	}
}

func (s Sql) apply(t string) string {
	if s.lowercase {
		return strings.ToLower(t)
	}

	return t
}

// GetDialect ...
func (s Sql) GetDialect() SqlDialect {
	return s.dialect
}

// IsMysql ...
func (s Sql) IsMysql() bool {
	return s.dialect == MysqlDialect
}

// IsPostgres ...
func (s Sql) IsPostgres() bool {
	return s.dialect == PostgresDialect
}

// IsSqlserver ...
func (s Sql) IsSqlserver() bool {
	return s.dialect == SqlserverDialect
}

// IsSqlite ...
func (s Sql) IsSqlite() bool {
	return s.dialect == SqliteDialect
}

// IsLowercase ...
func (s Sql) IsLowercase() bool {
	return s.lowercase
}

// CreateTableStm ...
func (s Sql) CreateTableStm() string {
	return s.apply("CREATE TABLE %s (\n%s\n)%s;")
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

// CreateTableMigration ...
func (s Sql) CreateTableMigration() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply(`CREATE TABLE IF NOT EXISTS %s (
 version    BIGINT PRIMARY KEY,
 dirty      BOOLEAN
);`)

	default:
		// TODO: mysql template is default for other dialects
		return s.apply(`CREATE TABLE IF NOT EXISTS %s (
 version    bigint(20) PRIMARY KEY,
 dirty      BOOLEAN
);`)
	}
}

// DropTableMigration ...
func (s Sql) DropTableMigration() string {
	return s.apply("DROP TABLE IF EXISTS %s;")
}

// InsertMigrationVersion ...
func (s Sql) InsertMigrationVersion() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("TRUNCATE TABLE %s;\nINSERT INTO %s (version, dirty) VALUES (%d, %t);")

	default:
		// TODO: mysql template is default for other dialects
		return s.apply("TRUNCATE %s;\nINSERT INTO %s (version, dirty) VALUES (%d, %t);")
	}
}

// RollbackMigrationVersion ...
func (s Sql) RollbackMigrationVersion() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("TRUNCATE TABLE %s;")

	default:
		// TODO: mysql template is default for other dialects
		return s.apply("TRUNCATE %s;")
	}
}
