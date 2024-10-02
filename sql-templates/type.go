package sql_templates

// BooleanType ...
func (s Sql) BooleanType() string {
	if s.IsSqlite() {
		return s.apply("INTEGER")
	}

	return s.apply("BOOLEAN")
}

// TinyIntType ...
func (s Sql) TinyIntType() string {
	return s.apply("TINYINT")
}

// SmallIntType ...
func (s Sql) SmallIntType() string {
	return s.apply("SMALLINT")
}

// IntType ...
func (s Sql) IntType() string {
	if s.IsSqlite() {
		return s.apply("INTEGER")
	}

	return s.apply("INT")
}

// BigIntType ...
func (s Sql) BigIntType() string {
	if s.IsSqlite() {
		return s.apply("INTEGER")
	}

	return s.apply("BIGINT")
}

// FloatType ...
func (s Sql) FloatType() string {
	if s.IsSqlite() {
		return s.apply("REAL")
	}

	return s.apply("FLOAT")
}

// DoubleType ...
func (s Sql) DoubleType() string {
	if s.IsSqlite() {
		return s.apply("REAL")
	}

	return s.apply("DOUBLE")
}

// TextType ...
func (s Sql) TextType() string {
	switch s.dialect {
	case SqliteDialect:
		return "TEXT"

	default:
		return s.apply("TEXT")
	}
}

// DatetimeType ...
func (s Sql) DatetimeType() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("TIMESTAMP")

	case SqliteDialect:
		return "TEXT" // TEXT as ISO8601 strings ("YYYY-MM-DD HH:MM:SS.SSS")

	default:
		return s.apply("DATETIME")
	}
}

// PointerType ...
func (s Sql) PointerType() string {
	return s.apply("POINTER")
}

// UnspecificType ...
func (s Sql) UnspecificType() string {
	return s.apply("UNSPECIFIED")
}
