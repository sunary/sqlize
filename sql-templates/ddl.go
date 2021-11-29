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
	if s.IsPostgres {
		return s.apply(`CREATE TABLE IF NOT EXISTS %s (
 id         SERIAL PRIMARY KEY,
 version    BIGINT,
 dirty      BOOLEAN,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 UNIQUE(version)
);`)
	}

	return s.apply(`CREATE TABLE IF NOT EXISTS %s (
 id         int(11) AUTO_INCREMENT PRIMARY KEY,
 version    bigint(20),
 dirty      BOOLEAN,
 created_at datetime DEFAULT CURRENT_TIMESTAMP(),
 updated_at datetime DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
 CONSTRAINT idx_version UNIQUE (version)
);`)
}

// DropTableMigration ...
func (s Sql) DropTableMigration() string {
	return s.apply("DROP TABLE IF EXISTS %s;")
}

// InsertMigrationVersion ...
func (s Sql) InsertMigrationVersion() string {
	return s.apply("INSERT INTO %s (version, dirty) VALUES (%d, %t);")
}

// RollbackMigrationVersion ...
func (s Sql) RollbackMigrationVersion() string {
	if s.IsPostgres {
		return s.apply("UPDATE %s SET dirty = FALSE, updated_at = CURRENT_TIMESTAMP WHERE version = %d;")
	}

	return s.apply("UPDATE %s SET dirty = FALSE WHERE version = %d;")
}
