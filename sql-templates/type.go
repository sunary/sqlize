package sql_templates

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
	return s.apply("DATETIME")
}

func (s Sql) PointerType() string {
	return s.apply("POINTER")
}

func (s Sql) UnspecificType() string {
	return s.apply("UNSPECIFIED")
}
