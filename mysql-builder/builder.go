package mysql_builder

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sunary/sqlize/mysql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	SqlTagDefault       = "sql"
	columnPrefix        = "column:"
	typePrefix          = "type:"
	defaultPrefix       = "default:"
	indexPrefix         = "index:"
	foreignKeyPrefix    = "foreignkey:"
	associationFkPrefix = "association_foreignkey:"
	isPrimaryKey        = "primary_key"
	isUnique            = "unique"
	isAutoIncrement     = "auto_increment"
	funcTableName       = "TableName"
)

type SqlBuilder struct {
	isLower bool
	sqlTag  string
}

func NewSqlBuilder(opts ...SqlBuilderOption) *SqlBuilder {
	o := sqlBuilderOptions{
		isLower: false,
		sqlTag:  SqlTagDefault,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	return &SqlBuilder{
		isLower: o.isLower,
		sqlTag:  o.sqlTag,
	}
}

func (s SqlBuilder) AddTable(tb interface{}) string {
	tableName := getTableName(tb)
	maxLen := 0

	fields := make([][]string, 0)
	indexes := make([]string, 0)
	v := reflect.ValueOf(tb)
	t := reflect.TypeOf(tb)
	for j := 0; j < t.NumField(); j++ {
		field := t.Field(j)
		gtag := field.Tag.Get(s.sqlTag)
		if gtag == "-" {
			continue
		}

		gts := strings.Split(gtag, ";")
		columnDeclare := utils.ToSnakeCase(field.Name)
		defaultDeclare := ""
		typeDeclare := ""
		isFkDeclare := false
		isPkDeclare := false
		isUniqueDeclare := false
		indexDeclare := ""
		indexColumns := ""
		isAutoDeclare := false
		for _, gt := range gts {
			gtLower := strings.ToLower(gt)
			if strings.HasPrefix(gtLower, foreignKeyPrefix) || strings.HasPrefix(gtLower, associationFkPrefix) {
				isFkDeclare = true
				break
			} else if strings.HasPrefix(gtLower, columnPrefix) {
				columnDeclare = gt[len(columnPrefix):]
			} else if strings.HasPrefix(gtLower, typePrefix) {
				typeDeclare = gt[len(typePrefix):]
			} else if strings.HasPrefix(gtLower, defaultPrefix) {
				defaultDeclare = fmt.Sprintf(mysql_templates.DefaultOption(s.isLower), gt[len(defaultPrefix):])
			} else if strings.HasPrefix(gtLower, indexPrefix) {
				indexDeclare = gt[len(indexPrefix):]
				if idxFields := strings.Split(indexDeclare, ","); len(idxFields) > 1 {
					indexDeclare = fmt.Sprintf("idx_%s", strings.Join(idxFields, "_"))
					indexColumns = strings.Join(utils.EscapeSqlNames(idxFields), ", ")
				} else {
					indexColumns = utils.EscapeSqlName(columnDeclare)
				}
			} else if gtLower == isPrimaryKey {
				isPkDeclare = true
			} else if gtLower == isUnique {
				isUniqueDeclare = true
			} else if gtLower == isAutoIncrement {
				isAutoDeclare = true
			}
		}

		if isFkDeclare {
			continue
		}

		if indexDeclare != "" {
			if isUniqueDeclare {
				indexes = append(indexes, fmt.Sprintf(mysql_templates.CreateUniqueIndexStm(s.isLower), utils.EscapeSqlName(indexDeclare), utils.EscapeSqlName(tableName), indexColumns))
			} else {
				indexes = append(indexes, fmt.Sprintf(mysql_templates.CreateIndexStm(s.isLower), utils.EscapeSqlName(indexDeclare), utils.EscapeSqlName(tableName), indexColumns))
			}
		}

		if len(columnDeclare) > maxLen {
			maxLen = len(columnDeclare)
		}

		fs := []string{columnDeclare}
		if typeDeclare != "" {
			fs = append(fs, typeDeclare)
		} else {
			fs = append(fs, s.sqlType(v.Field(j).Interface(), ""))
		}
		if defaultDeclare != "" {
			fs = append(fs, defaultDeclare)
		}
		if isAutoDeclare {
			fs = append(fs, mysql_templates.AutoIncrementOption(s.isLower))
		}
		if isPkDeclare {
			fs = append(fs, mysql_templates.PrimaryOption(s.isLower))
		}

		fields = append(fields, fs)
	}

	fs := make([]string, 0)
	for _, f := range fields {
		fs = append(fs, fmt.Sprintf("  %s%s%s", utils.EscapeSqlName(f[0]), strings.Repeat(" ", maxLen-len(f[0])+1), strings.Join(f[1:], " ")))
	}

	sql := []string{fmt.Sprintf(mysql_templates.CreateTableStm(s.isLower), utils.EscapeSqlName(tableName), strings.Join(fs, ",\n"))}
	sql = append(sql, indexes...)
	return strings.Join(sql, "\n")
}

func (s SqlBuilder) RemoveTable(tb interface{}) string {
	return fmt.Sprintf(mysql_templates.DropTableStm(s.isLower), utils.EscapeSqlName(getTableName(tb)))
}

func (s SqlBuilder) sqlType(v interface{}, suffix string) string {
	if reflect.ValueOf(v).Kind() == reflect.Ptr {
		vv := reflect.Indirect(reflect.ValueOf(v))
		if vv.IsZero() {
			return mysql_templates.UnspecificType(s.isLower)
		}

		return s.sqlType(vv.Interface(), mysql_templates.NullValue(s.isLower))
	}

	if suffix != "" {
		suffix = " " + suffix
	}

	switch v.(type) {
	case bool:
		return mysql_templates.BooleanType(s.isLower) + suffix

	case int8, uint8:
		return mysql_templates.TinyIntType(s.isLower) + suffix

	case int16, uint16:
		return mysql_templates.SmallIntType(s.isLower) + suffix

	case int, int32, uint32:
		return mysql_templates.IntType(s.isLower) + suffix

	case int64, uint64:
		return mysql_templates.BigIntType(s.isLower) + suffix

	case float32:
		return mysql_templates.FloatType(s.isLower) + suffix

	case float64:
		return mysql_templates.DoubleType(s.isLower) + suffix

	case string:
		return mysql_templates.TextType(s.isLower) + suffix

	case time.Time:
		return mysql_templates.DatetimeType(s.isLower) + suffix

	default:
		return mysql_templates.UnspecificType(s.isLower)
	}
}

func getTableName(t interface{}) string {
	st := reflect.TypeOf(t)
	if _, ok := st.MethodByName(funcTableName); ok {
		v := reflect.ValueOf(t).MethodByName(funcTableName).Call(nil)
		if len(v) > 0 {
			return v[0].String()
		}
	}

	name := ""
	if t := reflect.TypeOf(t); t.Kind() == reflect.Ptr {
		name = t.Elem().Name()
	} else {
		name = t.Name()
	}

	return utils.ToSnakeCase(name)
}
