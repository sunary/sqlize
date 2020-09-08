package mysql_templates

func PrimaryOption(isLower bool) string {
	return fromStr("PRIMARY KEY").apply(isLower)
}

func AutoIncrementOption(isLower bool) string {
	return fromStr("AUTO_INCREMENT").apply(isLower)
}

func DefaultOption(isLower bool) string {
	return fromStr("DEFAULT %s").apply(isLower)
}

func NullValue(isLower bool) string {
	return fromStr("NULL").apply(isLower)
}
