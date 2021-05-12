package sql_builder

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
	columnPrefix        = "column:"  // column:column_name
	oldNameMark         = ",old:"    // column:column_name,old:old_name
	typePrefix          = "type:"    // type:VARCHAR(64)
	defaultPrefix       = "default:" // default:0
	indexPrefix         = "index:"   // index:idx_name index:column1,column2
	foreignKeyPrefix    = "foreignkey:"
	associationFkPrefix = "association_foreignkey:"
	isPrimaryKey        = "primary_key"
	isUnique            = "unique"
	isAutoIncrement     = "auto_increment"
	funcTableName       = "TableName"
)

var (
	reflectValueFields = map[string]struct{}{
		"typ":  {},
		"ptr":  {},
		"flag": {},
	}
)

type SqlBuilder struct {
	isPostgres bool
	isLower    bool
	sqlTag     string
}

func NewSqlBuilder(opts ...SqlBuilderOption) *SqlBuilder {
	o := sqlBuilderOptions{
		isLower:    false,
		isPostgres: false,
		sqlTag:     SqlTagDefault,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	return &SqlBuilder{
		isPostgres: o.isPostgres,
		isLower:    o.isLower,
		sqlTag:     o.sqlTag,
	}
}

func (s SqlBuilder) AddTable(obj interface{}) string {
	tableName := getTableName(obj)
	columns, columnsHistory, indexes := s.parseStruct(tableName, obj)

	sqlPrimaryKey := mysql_templates.PrimaryOption(s.isLower)
	for i := range columns {
		if strings.Index(columns[i], sqlPrimaryKey) > 0 {
			columns[0], columns[i] = columns[i], columns[0]
			break
		}
	}

	sql := []string{fmt.Sprintf(mysql_templates.CreateTableStm(s.isLower), utils.EscapeSqlName(tableName), strings.Join(columns, ",\n"))}
	for _, h := range columnsHistory {
		sql = append(sql, fmt.Sprintf(mysql_templates.AlterTableRenameColumnStm(s.isLower), utils.EscapeSqlName(tableName), utils.EscapeSqlName(h[0]), utils.EscapeSqlName(h[1])))
	}

	sql = append(sql, indexes...)

	return strings.Join(sql, "\n")
}

func (s SqlBuilder) parseStruct(tableName string, obj interface{}) ([]string, [][2]string, []string) {
	maxLen := 0

	rawCols := make([][]string, 0)

	embedColumns := make([]string, 0)
	embedColumnsHistory := make([][2]string, 0)
	embedIndexes := make([]string, 0)

	columns := make([]string, 0)
	columnsHistory := make([][2]string, 0)
	indexes := make([]string, 0)

	primaryKey := ""
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)
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
			hasBreak := false

			gtLower := strings.ToLower(gt)
			switch {
			case strings.HasPrefix(gtLower, columnPrefix):
				columnNames := strings.Split(gt[len(columnPrefix):], oldNameMark)
				if len(columnNames) == 1 {
					columnDeclare = columnNames[0]
				} else {
					columnsHistory = append(columnsHistory, [2]string{columnNames[1], columnNames[0]})
					columnDeclare = columnNames[1]
				}

			case strings.HasPrefix(gtLower, foreignKeyPrefix) || strings.HasPrefix(gtLower, associationFkPrefix):
				isFkDeclare = true
				hasBreak = true

			case strings.HasPrefix(gtLower, typePrefix):
				typeDeclare = gt[len(typePrefix):]

			case strings.HasPrefix(gtLower, defaultPrefix):
				defaultDeclare = fmt.Sprintf(mysql_templates.DefaultOption(s.isLower), gt[len(defaultPrefix):])

			case strings.HasPrefix(gtLower, indexPrefix):
				indexDeclare = gt[len(indexPrefix):]
				if idxFields := strings.Split(indexDeclare, ","); len(idxFields) > 1 {
					indexDeclare = fmt.Sprintf("idx_%s", strings.Join(idxFields, "_"))
					indexColumns = strings.Join(utils.EscapeSqlNames(idxFields), ", ")
				} else {
					indexColumns = utils.EscapeSqlName(columnDeclare)
				}

			case strings.HasPrefix(gtLower, isPrimaryKey):
				if gtLower == isPrimaryKey {
					isPkDeclare = true
				} else {
					pkDeclare := gt[len(isPrimaryKey)+1:]
					idxFields := strings.Split(pkDeclare, ",")
					primaryKey = strings.Join(utils.EscapeSqlNames(idxFields), ", ")
				}

			case gtLower == isUnique:
				isUniqueDeclare = true

			case gtLower == isAutoIncrement:
				isAutoDeclare = true
			}

			if hasBreak {
				break
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

		col := []string{columnDeclare}
		if typeDeclare != "" {
			col = append(col, typeDeclare)
		} else {
			if _, ok := reflectValueFields[t.Field(j).Name]; ok {
				continue
			}

			strType, isStruct := s.sqlType(v.Field(j).Interface(), "")
			if isStruct {
				_columns, _columnsHistory, _indexes := s.parseStruct(tableName, v.Field(j).Interface())
				embedColumns = append(embedColumns, _columns...)
				embedColumnsHistory = append(embedColumnsHistory, _columnsHistory...)
				embedIndexes = append(embedIndexes, _indexes...)
				continue
			} else {
				col = append(col, strType)
			}
		}
		if defaultDeclare != "" {
			col = append(col, defaultDeclare)
		}
		if isAutoDeclare {
			col = append(col, mysql_templates.AutoIncrementOption(s.isLower))
		}
		if isPkDeclare {
			col = append(col, mysql_templates.PrimaryOption(s.isLower))
		}

		rawCols = append(rawCols, col)
	}

	for _, f := range rawCols {
		columns = append(columns, fmt.Sprintf("  %s%s%s", utils.EscapeSqlName(f[0]), strings.Repeat(" ", maxLen-len(f[0])+1), strings.Join(f[1:], " ")))
	}

	if len(primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("  %s (%s)", mysql_templates.PrimaryOption(s.isLower), primaryKey))
	}

	return append(columns, embedColumns...), append(columnsHistory, embedColumnsHistory...), append(indexes, embedIndexes...)
}

func (s SqlBuilder) RemoveTable(tb interface{}) string {
	return fmt.Sprintf(mysql_templates.DropTableStm(s.isLower), utils.EscapeSqlName(getTableName(tb)))
}

func (s SqlBuilder) sqlType(v interface{}, suffix string) (string, bool) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Ptr:
		vv := reflect.Indirect(reflect.ValueOf(v))
		if reflect.ValueOf(v).Pointer() == 0 || vv.IsZero() {
			return mysql_templates.PointerType(s.isLower), false
		}

		return s.sqlType(vv.Interface(), mysql_templates.NullValue(s.isLower))

	case reflect.Struct:
		if _, ok := v.(time.Time); ok {
			break
		}

		return "", true
	}

	if suffix != "" {
		suffix = " " + suffix
	}

	return s.sqlPrimitiveType(v, suffix), false
}

func (s SqlBuilder) sqlPrimitiveType(v interface{}, suffix string) string {
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
