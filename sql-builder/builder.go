package sql_builder

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sunary/sqlize/sql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	// SqlTagDefault ...
	SqlTagDefault       = "sql"
	columnPrefix        = "column:"         // 'column:column_name'
	embeddedPrefix      = "embeddedprefix:" // 'embeddedPrefix:base_'
	previousNamePrefix  = ",previous:"      // 'column:column_name,previous:old_name'
	typePrefix          = "type:"           // 'type:VARCHAR(64)'
	defaultPrefix       = "default:"        // 'default:0'
	isAutoIncrement     = "auto_increment"
	isPrimaryKey        = "primary_key"
	isUnique            = "unique"
	isIndex             = "index"       // 'index' (=> idx_column_name)
	indexPrefix         = "index:"      // 'index:idx_name' or 'index:col_name' (=> idx_col_name) or 'index:col1,col2' (=> idx_col1_col2)
	indexTypePrefix     = "index_type:" // 'index_type:btree' (default) or 'index_type:hash'
	foreignKeyPrefix    = "foreignkey:"
	associationFkPrefix = "association_foreignkey:"
	enumTag             = "enum"
	funcTableName       = "TableName"
)

var (
	reflectValueFields = map[string]struct{}{
		"typ":  {},
		"ptr":  {},
		"flag": {},
	}
	ignoredFieldComment = map[string]struct{}{
		"id":         {},
		"created_at": {},
		"updated_at": {},
		"deleted_at": {},
	}
	compileEnumValues = regexp.MustCompile("[\\s\"'()]+")
)

// SqlBuilder ...
type SqlBuilder struct {
	sql        *sql_templates.Sql
	isPostgres bool
	sqlTag     string
	hasComment bool
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
		hasComment: o.hasComment,
	}
}

// AddTable ...
func (s SqlBuilder) AddTable(obj interface{}) string {
	tableName := getTableName(obj)
	columns, columnsHistory, indexes := s.parseStruct(tableName, "", obj)

	sqlPrimaryKey := s.sql.PrimaryOption()
	for i := range columns {
		if strings.Index(columns[i], sqlPrimaryKey) > 0 {
			columns[0], columns[i] = columns[i], columns[0]
			break
		}
	}

	sqls := []string{fmt.Sprintf(s.sql.CreateTableStm(),
		utils.EscapeSqlName(s.isPostgres, tableName),
		strings.Join(columns, ",\n"))}
	for _, h := range columnsHistory {
		sqls = append(sqls,
			fmt.Sprintf(s.sql.AlterTableRenameColumnStm(),
				utils.EscapeSqlName(s.isPostgres, tableName),
				utils.EscapeSqlName(s.isPostgres, h[0]),
				utils.EscapeSqlName(s.isPostgres, h[1])))
	}

	sqls = append(sqls, indexes...)

	return strings.Join(sqls, "\n")
}

type attrs struct {
	Name         string
	Prefix       string
	Type         string
	Value        string
	IsFk         bool
	IsPk         bool
	IsUnique     bool
	Index        string
	IndexType    string
	IndexColumns string
	IsAutoIncr   bool
	Comment      string
}

// parseStruct return columns, columnsHistory, indexes
func (s SqlBuilder) parseStruct(tableName, prefix string, obj interface{}) ([]string, [][2]string, []string) {
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

		at := attrs{
			Name: prefix + utils.ToSnakeCase(field.Name),
		}

		gts := strings.Split(gtag, ";")
		for _, gt := range gts {
			hasBreak := false

			gtLower := strings.ToLower(gt)
			switch {
			case strings.HasPrefix(gtLower, columnPrefix):
				columnNames := strings.Split(gt[len(columnPrefix):], previousNamePrefix)
				if len(columnNames) == 1 {
					at.Name = columnNames[0]
				} else {
					columnsHistory = append(columnsHistory, [2]string{columnNames[1], columnNames[0]})
					at.Name = columnNames[1]
				}
				at.Name = prefix + at.Name

			case strings.HasPrefix(gtLower, embeddedPrefix):
				at.Prefix = gt[len(embeddedPrefix):]

			case strings.HasPrefix(gtLower, foreignKeyPrefix) || strings.HasPrefix(gtLower, associationFkPrefix):
				at.IsFk = true
				hasBreak = true

			case strings.HasPrefix(gtLower, typePrefix):
				at.Type = gt[len(typePrefix):]
				if s.hasComment && strings.HasPrefix(gtLower[len(typePrefix):], enumTag) {
					enumRaw := gt[len(typePrefix)+len(enumTag):]
					enumStr := compileEnumValues.ReplaceAllString(enumRaw, "")
					at.Comment = createComment(at.Name, strings.Split(enumStr, ","))
				}

			case strings.HasPrefix(gtLower, defaultPrefix):
				at.Value = fmt.Sprintf(s.sql.DefaultOption(), gt[len(defaultPrefix):])

			case gtLower == isIndex:
				at.Index = createIndexName("", []string{at.Name}, at.Name)
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case strings.HasPrefix(gtLower, indexPrefix):
				idxFields := strings.Split(gt[len(indexPrefix):], ",")
				at.Index = createIndexName(prefix, idxFields, at.Name)
				if len(idxFields) > 1 {
					at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
				} else {
					at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)
				}

			case strings.HasPrefix(gtLower, indexTypePrefix):
				at.IndexType = gt[len(indexTypePrefix):]

			case strings.HasPrefix(gtLower, isPrimaryKey):
				if gtLower == isPrimaryKey {
					at.IsPk = true
				} else {
					pkDeclare := gt[len(isPrimaryKey)+1:]
					idxFields := strings.Split(pkDeclare, ",")
					primaryKey = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
				}

			case gtLower == isUnique:
				at.IsUnique = true

			case gtLower == isAutoIncrement:
				at.IsAutoIncr = true
			}

			if hasBreak {
				break
			}
		}

		if at.IsFk {
			continue
		}

		if at.Index != "" {
			var strIndex string
			if at.IsUnique {
				strIndex = fmt.Sprintf(s.sql.CreateUniqueIndexStm(at.IndexType),
					utils.EscapeSqlName(s.isPostgres, at.Index),
					utils.EscapeSqlName(s.isPostgres, tableName), at.IndexColumns)
			} else {
				strIndex = fmt.Sprintf(s.sql.CreateIndexStm(at.IndexType),
					utils.EscapeSqlName(s.isPostgres, at.Index),
					utils.EscapeSqlName(s.isPostgres, tableName), at.IndexColumns)
			}

			indexes = append(indexes, strIndex)
		}

		if len(at.Name) > maxLen {
			maxLen = len(at.Name)
		}

		col := []string{at.Name}
		if at.Type != "" {
			col = append(col, at.Type)
		} else {
			if _, ok := reflectValueFields[t.Field(j).Name]; ok {
				continue
			}

			strType, isStruct := s.sqlType(v.Field(j).Interface(), "")
			if isStruct {
				_columns, _columnsHistory, _indexes := s.parseStruct(tableName, prefix+at.Prefix, v.Field(j).Interface())
				embedColumns = append(embedColumns, _columns...)
				embedColumnsHistory = append(embedColumnsHistory, _columnsHistory...)
				embedIndexes = append(embedIndexes, _indexes...)
				continue
			} else {
				col = append(col, strType)
			}
		}

		if at.Value != "" {
			col = append(col, at.Value)
		}

		if at.IsAutoIncr {
			if s.sql.IsPostgres {
				col = []string{col[0], s.sql.AutoIncrementOption()}
			} else {
				col = append(col, s.sql.AutoIncrementOption())
			}
		}

		if at.IsPk {
			col = append(col, s.sql.PrimaryOption())
		}

		if s.hasComment {
			if at.Comment == "" {
				at.Comment = createComment(at.Name, nil)
			}

			if at.Comment != "" {
				col = append(col, fmt.Sprintf(s.sql.Comment(), at.Comment))
			}
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

// createIndexName format idx_field_names
func createIndexName(prefix string, indexColumns []string, column string) string {
	cols := make([]string, len(indexColumns))
	for i := range indexColumns {
		cols[i] = prefix + indexColumns[i]
	}

	if len(indexColumns) == 1 && indexColumns[0] != column {
		return indexColumns[0]
	}

	return fmt.Sprintf("idx_%s", strings.Join(indexColumns, "_"))
}

// createComment by enum values or field name
func createComment(column string, enums []string) string {
	if _, ok := ignoredFieldComment[column]; ok {
		return ""
	}

	if len(enums) > 0 {
		return fmt.Sprintf("enum values: %s", strings.Join(enums, ", "))
	}

	return strings.Replace(column, "_", " ", -1)
}

// prefix return sqlType, isStruct
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

// sqlNullType return sqlNullType, isPrimitiveType
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

// sqlPrimitiveType ...
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

// getTableName ...
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
