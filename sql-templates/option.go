package sql_templates

func (s Sql) PrimaryOption() string {
	return s.apply("PRIMARY KEY")
}

func (s Sql) AutoIncrementOption() string {
	return s.apply("AUTO_INCREMENT")
}

func (s Sql) DefaultOption() string {
	return s.apply("DEFAULT %s")
}

func (s Sql) NullValue() string {
	return s.apply("NULL")
}
