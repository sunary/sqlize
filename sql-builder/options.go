package sql_builder

type sqlBuilderOptions struct {
	isPostgres      bool
	isLower         bool
	sqlTag          string
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
		o.isPostgres = false
	})
}

// WithPostgresql default is mysql
func WithPostgresql() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isPostgres = true
	})
}

// WithSqlUppercase default
func WithSqlUppercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isLower = false
	})
}

// WithSqlLowercase default is uppercase
func WithSqlLowercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isLower = true
	})
}

// Table name plus s default
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
