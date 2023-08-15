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
	if s.IsPostgres() {
		return s.apply("SERIAL")
	}
	return s.apply("AUTO_INCREMENT")
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
