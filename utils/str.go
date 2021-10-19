package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	defaultMigrationTimeFormat = "20060102150405"
	// MigrationUpSuffix ...
	MigrationUpSuffix = ".up.sql"
	// MigrationDownSuffix ...
	MigrationDownSuffix = ".down.sql"
)

// MigrationFileName ...
func MigrationFileName(name string) string {
	re, _ := regexp.Compile(`[^\w\d\s-_]`)
	name = strings.ToLower(re.ReplaceAllString(name, ""))
	name = strings.Replace(name, "  ", " ", -1)
	name = strings.Replace(name, " ", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	return fmt.Sprintf("%s_%s", time.Now().Format(defaultMigrationTimeFormat), name)
}

// ToSnakeCase ...
func ToSnakeCase(input string) string {
	var sb strings.Builder
	var upperCount int
	for i, c := range input {
		switch {
		case isUppercase(c):
			if i > 0 && (upperCount == 0 || nextIsLower(input, i)) {
				sb.WriteByte('_')
			}
			sb.WriteByte(byte(c - 'A' + 'a'))
			upperCount++

		case isLowercase(c):
			sb.WriteByte(byte(c))
			upperCount = 0

		case isDigit(c):
			if i == 0 {
				panic("Identifier must start with a character: `" + input + "`")
			}
			sb.WriteByte(byte(c))

		default:
			panic("Invalid identifier: `" + input + "`")
		}
	}

	return sb.String()
}

// nextIsLower The next character is lower case, but not the last 's'.
// nextIsLower("HTMLFile", 1) expected: "html_file"
// => true
// nextIsLower("URLs", -1) expected: "urls"
// => false
func nextIsLower(input string, i int) bool {
	i++
	if i >= len(input) {
		return false
	}

	c := input[i]
	if c == 's' && i == len(input)-1 {
		return false
	}

	return isLowercase(rune(c))
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLowercase(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUppercase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// EscapeSqlName ...
func EscapeSqlName(isPostgres bool, name string) string {
	if isPostgres || name == "" {
		return name
	}

	return fmt.Sprintf("`%s`", strings.Trim(name, "`"))
}

// EscapeSqlNames ...
func EscapeSqlNames(isPostgres bool, names []string) []string {
	ns := make([]string, len(names))
	for i := range names {
		ns[i] = EscapeSqlName(isPostgres, names[i])
	}

	return ns
}
