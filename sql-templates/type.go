package sql_templates

var familyName = map[int32]string{
	0:   "bool",
	1:   "int",
	2:   "float",
	3:   "decimal",
	4:   "date",
	5:   "timestamp",
	6:   "interval",
	7:   "string",
	8:   "bytes",
	9:   "timestamptz",
	10:  "collated string",
	12:  "oid",
	13:  "unknown",
	14:  "uuid",
	15:  "array",
	16:  "inet",
	17:  "time",
	18:  "json",
	19:  "timetz",
	20:  "tuple",
	21:  "bit",
	100: "any",
}

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
	return s.apply("TEXT")
}

// DatetimeType ...
func (s Sql) DatetimeType() string {
	switch s.dialect {
	case PostgresDialect:
		return s.apply("TIMESTAMP")

	case SqliteDialect:
		return s.apply("TEXT") // TEXT as ISO8601 strings ("YYYY-MM-DD HH:MM:SS.SSS")

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

// FamilyName ...
func (s Sql) FamilyName(f int32) string {
	return familyName[f]
}
