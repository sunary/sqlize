package sql_builder

type sqlBuilderOptions struct {
	isPostgres bool
	isLower    bool
	sqlTag     string
	hasComment bool
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

// WithComment default is off
func WithComment() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.hasComment = true
	})
}
