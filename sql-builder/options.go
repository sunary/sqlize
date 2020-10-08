package sql_builder

type sqlBuilderOptions struct {
	isPostgres bool
	isLower    bool
	sqlTag     string
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

type SqlBuilderOption interface {
	apply(*sqlBuilderOptions)
}

func WithSqlTag(sqlTag string) SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.sqlTag = sqlTag
	})
}

func WithPostgresql() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isPostgres = true
	})
}

func WithSqlUppercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isLower = false
	})
}

func WithSqlLowercase() SqlBuilderOption {
	return newFuncSqlBuilderOption(func(o *sqlBuilderOptions) {
		o.isLower = false
	})
}
