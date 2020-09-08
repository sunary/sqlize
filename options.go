package sqlize

type sqlizeOptions struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string

	isLower bool
	sqlTag  string
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

type SqlizeOption interface {
	apply(*sqlizeOptions)
}

func WithMigrationFolder(path string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationFolder = path
	})
}

func WithMigrationSuffix(upSuffix, downSuffix string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationUpSuffix = upSuffix
		o.migrationDownSuffix = downSuffix
	})
}

func WithSqlTag(sqlTag string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.sqlTag = sqlTag
	})
}

func WithSqlUppercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.isLower = false
	})
}

func WithSqlLowercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.isLower = false
	})
}
