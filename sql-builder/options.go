package sql_builder

import (
	sql_templates "github.com/sunary/sqlize/sql-templates"
)

type sqlBuilderOptions struct {
	sqlTag          string
	dialect         sql_templates.SqlDialect
	lowercase       bool
	generateComment bool
	pluralTableName bool
}

type funcSqlBuilderOption struct {
	f func(*sqlBuilderOptions)
}

func (fso *funcSqlBuilderOption) apply(do *sqlBuilderOptions) {
	fso.f(do)
}

func newFuncSqlBuilderOption(f func(*sqlBuilderOptions)) *funcSqlBuilderOption {
	return &funcSqlBuilderOption{
		f: f,
	}
}

// SqlBuilderOption ...
type SqlBuilderOption interface {
	apply(*sqlBuilderOptions)
}

// WithSqlTag default tag is `sql`
func WithSqlTag(sqlTag string) SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.sqlTag = sqlTag
	})
}

// WithMysql default
func WithMysql() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.dialect = sql_templates.MysqlDialect
	})
}

// WithPostgresql ...
func WithPostgresql() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.dialect = sql_templates.PostgresDialect
	})
}

// WithSqlserver ...
func WithSqlserver() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.dialect = sql_templates.SqlserverDialect
	})
}

// WithSqlite ...
func WithSqlite() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.dialect = sql_templates.SqliteDialect
	})
}

// WithDialect ...
func WithDialect(dialect sql_templates.SqlDialect) SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.dialect = dialect
	})
}

// WithSqlUppercase default
func WithSqlUppercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.lowercase = false
	})
}

// WithSqlLowercase ...
func WithSqlLowercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.lowercase = true
	})
}

// WithPluralTableName Table name in plural convention - ending with `s`
func WithPluralTableName() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.pluralTableName = true
	})
}

// WithCommentGenerate default is off
func WithCommentGenerate() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.generateComment = true
	})
}
