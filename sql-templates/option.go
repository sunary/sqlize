package sql_templates

type SqlDialect string

const (
	MysqlDialect     = "mysql"
	PostgresDialect  = "postgres"
	SqlserverDialect = "sqlserver"
	SqliteDialect    = "sqlite3"
)

// PrimaryOption ...
func (s Sql) PrimaryOption() string {
	return s.apply("PRIMARY KEY")
}

// AutoIncrementOption ...
func (s Sql) AutoIncrementOption() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("SERIAL")

	default:
		// TODO: mysql template is default for other dialects
		return s.apply("AUTO_INCREMENT")
	}
}

// DefaultOption ...
func (s Sql) DefaultOption() string {
	return s.apply("DEFAULT %s")
}

// NotNullValue ...
func (s Sql) NotNullValue() string {
	return s.apply("NOT NULL")
}

// NullValue ...
func (s Sql) NullValue() string {
	return s.apply("NULL")
}

// NullValue ...
func (s Sql) Comment() string {
	return s.apply("COMMENT '%s'")
}
