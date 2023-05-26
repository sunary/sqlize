package sqlize

type sqlizeOptions struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	migrationTable      string

	isPostgres      bool
	isLower         bool
	sqlTag          string
	pluralTableName bool
	generateComment bool
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
		o.isLower = true
	})
}

// Table name plus s default
func WithTableAdds() SqlizeOption {
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
