package sqlize

import (
	sql_templates "github.com/sunary/sqlize/sql-templates"
)

type sqlizeOptions struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	migrationTable      string

	sqlTag           string
	dialect          sql_templates.SqlDialect
	lowercase        bool
	pluralTableName  bool
	generateComment  bool
	ignoreFieldOrder bool
}

type funcSqlizeOption struct {
	f func(*sqlizeOptions)
}

func (fmo *funcSqlizeOption) apply(do *sqlizeOptions) {
	fmo.f(do)
}

func newFuncSqlizeOption(f func(*sqlizeOptions)) *funcSqlizeOption {
	return &funcSqlizeOption{
		f: f,
	}
}

// SqlizeOption ...
type SqlizeOption interface {
	apply(*sqlizeOptions)
}

// WithMigrationFolder ...
func WithMigrationFolder(path string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationFolder = path
	})
}

// WithMigrationSuffix ...
func WithMigrationSuffix(upSuffix, downSuffix string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationUpSuffix = upSuffix
		o.migrationDownSuffix = downSuffix
	})
}

// WithMigrationTable default is 'schema_migration'
func WithMigrationTable(table string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationTable = table
	})
}

// WithSqlTag default is `sql`
func WithSqlTag(sqlTag string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.sqlTag = sqlTag
	})
}

// WithMysql default
func WithMysql() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.dialect = sql_templates.MysqlDialect
	})
}

// WithPostgresql ...
func WithPostgresql() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.dialect = sql_templates.PostgresDialect
	})
}

// WithSqlserver ...
func WithSqlserver() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.dialect = sql_templates.SqlserverDialect
	})
}

// WithSqlite ...
func WithSqlite() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.dialect = sql_templates.SqliteDialect
	})
}

// WithSqlUppercase default
func WithSqlUppercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.lowercase = false
	})
}

// WithSqlLowercase ...
func WithSqlLowercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.lowercase = true
	})
}

// Table name plus s default
func WithPluralTableName() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.pluralTableName = true
	})
}

// WithCommentGenerate default is off
func WithCommentGenerate() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.generateComment = true
	})
}

// WithIgnoreFieldOrder ...
func WithIgnoreFieldOrder() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.ignoreFieldOrder = true
	})
}
