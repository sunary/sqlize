package mysql_templates

func BooleanType(isLower bool) string {
	return fromStr("BOOLEAN").apply(isLower)
}

func TinyIntType(isLower bool) string {
	return fromStr("TINYINT").apply(isLower)
}

func SmallIntType(isLower bool) string {
	return fromStr("SMALLINT").apply(isLower)
}

func IntType(isLower bool) string {
	return fromStr("INT").apply(isLower)
}

func BigIntType(isLower bool) string {
	return fromStr("BIGINT").apply(isLower)
}

func FloatType(isLower bool) string {
	return fromStr("FLOAT").apply(isLower)
}

func DoubleType(isLower bool) string {
	return fromStr("DOUBLE").apply(isLower)
}

func TextType(isLower bool) string {
	return fromStr("TEXT").apply(isLower)
}
func DatetimeType(isLower bool) string {
	return fromStr("DATETIME").apply(isLower)
}

func PointerType(isLower bool) string {
	return fromStr("POINTER").apply(isLower)
}

func UnspecificType(isLower bool) string {
	return fromStr("UNSPECIFIED").apply(isLower)
}
