package sql_builder

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sunary/sqlize/sql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	SqlTagDefault       = "sql"
	columnPrefix        = "column:"     // column:column_name
	oldNameMark         = ",old:"       // column:column_name,old:old_name
	typePrefix          = "type:"       // type:VARCHAR(64)
	defaultPrefix       = "default:"    // default:0
	indexPrefix         = "index:"      // index:idx_name or index:col1,col2 (=> idx_col1_col2)
	indexTypePrefix     = "index_type:" // index_type:btree (default) or index_type:hash
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

// SqlBuilder ...
type SqlBuilder struct {
	sql        *sql_templates.Sql
	isPostgres bool
	sqlTag     string
}

// NewSqlBuilder ...
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
		sql:        sql_templates.NewSql(o.isPostgres, o.isLower),
		isPostgres: o.isPostgres,
		sqlTag:     o.sqlTag,
	}
}

// AddTable ...
func (s SqlBuilder) AddTable(obj interface{}) string {
	tableName := getTableName(obj)
	columns, columnsHistory, indexes := s.parseStruct(tableName, obj)

	sqlPrimaryKey := s.sql.PrimaryOption()
	for i := range columns {
		if strings.Index(columns[i], sqlPrimaryKey) > 0 {
			columns[0], columns[i] = columns[i], columns[0]
			break
		}
	}

	sql := []string{fmt.Sprintf(s.sql.CreateTableStm(),
		utils.EscapeSqlName(s.isPostgres, tableName),
		strings.Join(columns, ",\n"))}
	for _, h := range columnsHistory {
		sql = append(sql,
			fmt.Sprintf(s.sql.AlterTableRenameColumnStm(),
				utils.EscapeSqlName(s.isPostgres, tableName),
				utils.EscapeSqlName(s.isPostgres, h[0]),
				utils.EscapeSqlName(s.isPostgres, h[1])))
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
		indexTypeDeclare := ""
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
				defaultDeclare = fmt.Sprintf(s.sql.DefaultOption(), gt[len(defaultPrefix):])

			case strings.HasPrefix(gtLower, indexPrefix):
				indexDeclare = gt[len(indexPrefix):]
				if idxFields := strings.Split(indexDeclare, ","); len(idxFields) > 1 {
					indexDeclare = fmt.Sprintf("idx_%s", strings.Join(idxFields, "_"))
					indexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
				} else {
					indexColumns = utils.EscapeSqlName(s.isPostgres, columnDeclare)
				}
			case strings.HasPrefix(gtLower, indexTypePrefix):
				indexTypeDeclare = gt[len(indexTypePrefix):]

			case strings.HasPrefix(gtLower, isPrimaryKey):
				if gtLower == isPrimaryKey {
					isPkDeclare = true
				} else {
					pkDeclare := gt[len(isPrimaryKey)+1:]
					idxFields := strings.Split(pkDeclare, ",")
					primaryKey = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
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
			var strIndex string
			if isUniqueDeclare {
				strIndex = fmt.Sprintf(s.sql.CreateUniqueIndexStm(indexTypeDeclare),
					utils.EscapeSqlName(s.isPostgres, indexDeclare),
					utils.EscapeSqlName(s.isPostgres, tableName), indexColumns)
			} else {
				strIndex = fmt.Sprintf(s.sql.CreateIndexStm(indexTypeDeclare),
					utils.EscapeSqlName(s.isPostgres, indexDeclare),
					utils.EscapeSqlName(s.isPostgres, tableName), indexColumns)
			}

			indexes = append(indexes, strIndex)
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
			if s.sql.IsPostgres {
				col = []string{col[0], s.sql.AutoIncrementOption()}
			} else {
				col = append(col, s.sql.AutoIncrementOption())
			}
		}
		if isPkDeclare {
			col = append(col, s.sql.PrimaryOption())
		}

		rawCols = append(rawCols, col)
	}

	for _, f := range rawCols {
		columns = append(columns, fmt.Sprintf("  %s%s%s", utils.EscapeSqlName(s.isPostgres, f[0]), strings.Repeat(" ", maxLen-len(f[0])+1), strings.Join(f[1:], " ")))
	}

	if len(primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("  %s (%s)", s.sql.PrimaryOption(), primaryKey))
	}

	return append(columns, embedColumns...), append(columnsHistory, embedColumnsHistory...), append(indexes, embedIndexes...)
}

// RemoveTable ...
func (s SqlBuilder) RemoveTable(tb interface{}) string {
	return fmt.Sprintf(s.sql.DropTableStm(), utils.EscapeSqlName(s.isPostgres, getTableName(tb)))
}

func (s SqlBuilder) sqlType(v interface{}, suffix string) (string, bool) {
	if t, ok := s.sqlNullType(v, s.sql.NullValue()); ok {
		return t, false
	}

	switch reflect.ValueOf(v).Kind() {
	case reflect.Ptr:
		vv := reflect.Indirect(reflect.ValueOf(v))
		if reflect.ValueOf(v).Pointer() == 0 || vv.IsZero() {
			return s.sql.PointerType(), false
		}

		return s.sqlType(vv.Interface(), s.sql.NullValue())

	case reflect.Struct:
		if _, ok := v.(time.Time); ok {
			break
		}

		return "", true
	}

	return s.sqlPrimitiveType(v, suffix), false
}

func (s SqlBuilder) sqlNullType(v interface{}, suffix string) (string, bool) {
	if suffix != "" {
		suffix = " " + suffix
	}

	switch v.(type) {
	case sql.NullBool:
		return s.sql.BooleanType() + suffix, true

	case sql.NullInt32:
		return s.sql.IntType() + suffix, true

	case sql.NullInt64:
		return s.sql.BigIntType() + suffix, true

	case sql.NullFloat64:
		return s.sql.DoubleType() + suffix, true

	case sql.NullString:
		return s.sql.TextType() + suffix, true

	case sql.NullTime:
		return s.sql.DatetimeType() + suffix, true

	default:
		return "", false
	}
}

func (s SqlBuilder) sqlPrimitiveType(v interface{}, suffix string) string {
	if suffix != "" {
		suffix = " " + suffix
	}

	switch v.(type) {
	case bool:
		return s.sql.BooleanType() + suffix

	case int8, uint8:
		return s.sql.TinyIntType() + suffix

	case int16, uint16:
		return s.sql.SmallIntType() + suffix

	case int, int32, uint32:
		return s.sql.IntType() + suffix

	case int64, uint64:
		return s.sql.BigIntType() + suffix

	case float32:
		return s.sql.FloatType() + suffix

	case float64:
		return s.sql.DoubleType() + suffix

	case string:
		return s.sql.TextType() + suffix

	case time.Time:
		return s.sql.DatetimeType() + suffix

	default:
		return s.sql.UnspecificType()
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
