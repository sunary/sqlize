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

func (s Sql) BooleanType() string {
	return s.apply("BOOLEAN")
}

func (s Sql) TinyIntType() string {
	return s.apply("TINYINT")
}

func (s Sql) SmallIntType() string {
	return s.apply("SMALLINT")
}

func (s Sql) IntType() string {
	return s.apply("INT")
}

func (s Sql) BigIntType() string {
	return s.apply("BIGINT")
}

func (s Sql) FloatType() string {
	return s.apply("FLOAT")
}

func (s Sql) DoubleType() string {
	return s.apply("DOUBLE")
}

func (s Sql) TextType() string {
	return s.apply("TEXT")
}

func (s Sql) DatetimeType() string {
	if s.IsPostgres {
		return s.apply("TIMESTAMP")
	}

	return s.apply("DATETIME")
}

func (s Sql) PointerType() string {
	return s.apply("POINTER")
}

func (s Sql) UnspecificType() string {
	return s.apply("UNSPECIFIED")
}

func (s Sql) FamilyName(f int32) string {
	return familyName[f]
}
