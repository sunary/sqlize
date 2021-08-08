package sqlize

type sqlizeOptions struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string

	isPostgres bool
	isLower    bool
	sqlTag     string
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

// WithMigrationSuffix default
func WithMigrationSuffix(upSuffix, downSuffix string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.migrationUpSuffix = upSuffix
		o.migrationDownSuffix = downSuffix
	})
}

// WithPostgresql default is mysql
func WithPostgresql() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.isPostgres = true
	})
}

// WithSqlTag default is `sql`
func WithSqlTag(sqlTag string) SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.sqlTag = sqlTag
	})
}

// WithSqlUppercase default
func WithSqlUppercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.isLower = false
	})
}

// WithSqlLowercase default is uppercase
func WithSqlLowercase() SqlizeOption {
	return newFuncSqlizeOption(func(o *sqlizeOptions) {
		o.isLower = false
	})
}
