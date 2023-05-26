package sql_builder

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	sql_templates "github.com/sunary/sqlize/sql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	// SqlTagDefault ...
	SqlTagDefault      = "sql"
	prefixColumn       = "column:"          // set column name, eg: 'column:column_name'
	prefixEmbedded     = "embedded_prefix:" // set embed prefix for flatten struct, eg: 'embedded_prefix:base_'
	prefixPreviousName = ",previous:"       // mark previous name-field, eg: 'column:column_name,previous:old_name'
	prefixType         = "type:"            // set field type, eg: 'type:VARCHAR(64)'
	prefixDefault      = "default:"         // set default value, eg: 'default:0'
	prefixComment      = "comment:"         // comment field, eg: 'comment:sth you want to comment'

	tagIsNull          = "null"
	tagIsNotNull       = "not_null"
	tagIsAutoIncrement = "auto_increment"
	tagIsPrimaryKey    = "primary_key" // this field is primary key, eg: 'primary_key'
	tagIsIndex         = "index"       // indexing this field, eg: 'index' (=> idx_column_name)
	tagIsUniqueIndex   = "unique"      // unique indexing this field, eg: 'unique' (=> idx_column_name)

	prefixIndex        = "index:"         // indexing with name, eg: 'index:idx_name'
	prefixUniqueIndex  = "unique:"        // unique indexing with name, eg: 'unique:idx_name'
	prefixIndexColumns = "index_columns:" // indexing these fields, eg: 'index_columns:col1,col2' (=> idx_col1_col2)
	prefixIndexType    = "index_type:"    // indexing with type, eg: 'index_type:btree' (default) or 'index_type:hash'

	prefixForeignKey    = "foreign_key:"             // 'foreign_key:'
	prefixAssociationFk = "association_foreign_key:" // 'association_foreign_key:'
	tagEnum             = "enum"
	tagIsSquash         = "squash"
	tagIsEmbedded       = "embedded"
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
	sql             *sql_templates.Sql
	isPostgres      bool
	sqlTag          string
	generateComment bool
	isTableAdds     bool
}

// NewSqlBuilder ...
func NewSqlBuilder(opts ...SqlBuilderOption) *SqlBuilder {
	o := sqlBuilderOptions{
		isLower:     false,
		isPostgres:  false,
		isTableAdds: false,
		sqlTag:      SqlTagDefault,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	return &SqlBuilder{
		sql:             sql_templates.NewSql(o.isPostgres, o.isLower),
		isPostgres:      o.isPostgres,
		sqlTag:          o.sqlTag,
		isTableAdds:     o.isTableAdds,
		generateComment: o.generateComment,
	}
}

// AddTable ...
func (s SqlBuilder) AddTable(obj interface{}) string {
	tableName := s.getTableName(obj)
	columns, columnsHistory, indexes := s.parseStruct(tableName, "", obj)

	sqlPrimaryKey := s.sql.PrimaryOption()
	for i := range columns {
		if strings.Index(columns[i], sqlPrimaryKey) > 0 {
			columns[0], columns[i] = columns[i], columns[0]
			break
		}
	}

	tableComment := ""
	if s.generateComment {
		tableComment = fmt.Sprintf(` COMMENT "%s"`, tableName)
	}
	sqls := []string{fmt.Sprintf(s.sql.CreateTableStm(),
		utils.EscapeSqlName(s.isPostgres, tableName),
		strings.Join(columns, ",\n"), tableComment)}
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
	IsPk         bool
	IsFk         bool
	IsUnique     bool
	Index        string
	IndexType    string
	IndexColumns string
	IsNull       bool
	IsNotNull    bool
	IsAutoIncr   bool
	Comment      string
	IsEmbedded   bool
}

/**
 * 驼峰转蛇形 snake string
 * @description XxYy to xx_yy , XxYY to xx_y_y
 * @param s 需要转换的字符串
 * @return string
 **/
func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90（A-Z），97-122（a-z）
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	//ToLower把大写字母统一转小写
	return strings.ToLower(string(data[:]))
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

	var pkFields []string
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
			gtLower := ""
			if strings.HasSuffix(gt, "Key") || strings.Contains(gt, "Key:") {
				gtLower = snakeString(gt)
			} else {
				gtLower = strings.ToLower(gt)
			}
			switch {
			case strings.HasPrefix(gtLower, prefixColumn):
				columnNames := strings.Split(gt[len(prefixColumn):], prefixPreviousName)
				if len(columnNames) == 1 {
					at.Name = columnNames[0]
				} else {
					columnsHistory = append(columnsHistory, [2]string{columnNames[1], columnNames[0]})
					at.Name = columnNames[1]
				}
				at.Name = prefix + at.Name

			case strings.HasPrefix(gtLower, prefixEmbedded):
				at.Prefix = gt[len(prefixEmbedded):]
				at.IsEmbedded = true

			case strings.HasPrefix(gtLower, prefixForeignKey) || strings.HasPrefix(gtLower, prefixAssociationFk):
				at.IsFk = true
				hasBreak = true

			case strings.HasPrefix(gtLower, prefixType):
				at.Type = gt[len(prefixType):]
				if s.generateComment && at.Comment == "" && strings.HasPrefix(gtLower[len(prefixType):], tagEnum) {
					enumRaw := gt[len(prefixType)+len(tagEnum):]
					enumStr := compileEnumValues.ReplaceAllString(enumRaw, "")
					at.Comment = createCommentFromEnum(strings.Split(enumStr, ","))
				}

			case strings.HasPrefix(gtLower, prefixDefault):
				at.Value = fmt.Sprintf(s.sql.DefaultOption(), gt[len(prefixDefault):])

			case strings.HasPrefix(gtLower, prefixComment):
				at.Comment = gt[len(prefixComment):]

			case gtLower == tagIsPrimaryKey:
				at.IsPk = true

			case gtLower == tagIsIndex:
				at.Index = getWhenEmpty(at.Index, createIndexName("", nil, at.Name))
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case gtLower == tagIsUniqueIndex:
				at.IsUnique = true
				at.Index = getWhenEmpty(at.Index, createIndexName("", nil, at.Name))
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case strings.HasPrefix(gtLower, prefixIndex):
				idxFields := strings.Split(gt[len(prefixIndex):], ",")
				at.Index = createIndexName(prefix, idxFields, at.Name)

				if len(idxFields) > 1 {
					at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
				} else {
					at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)
				}

			case strings.HasPrefix(gtLower, prefixUniqueIndex):
				at.IsUnique = true
				at.Index = createIndexName(prefix, []string{gt[len(prefixUniqueIndex):]}, at.Name)
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case strings.HasPrefix(gtLower, prefixIndexColumns):
				idxFields := strings.Split(gt[len(prefixIndexColumns):], ",")
				if at.IsPk {
					pkFields = idxFields
				}

				at.Index = createIndexName(prefix, idxFields, at.Name)
				at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")

			case strings.HasPrefix(gtLower, prefixIndexType):
				at.Index = getWhenEmpty(at.Index, createIndexName(prefix, nil, at.Name))

				if len(at.IndexColumns) == 0 {
					at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, []string{at.Name}), ", ")
				}
				at.IndexType = gt[len(prefixIndexType):]

			case gtLower == tagIsNull:
				at.IsNotNull = true

			case gtLower == tagIsNotNull:
				at.IsNotNull = true

			case gtLower == tagIsAutoIncrement:
				at.IsAutoIncr = true

			case gtLower == tagIsSquash, gtLower == tagIsEmbedded:
				at.IsEmbedded = true
			}

			if hasBreak {
				break
			}
		}

		if at.IsFk {
			continue
		}

		if at.IsPk {
			// create primary key multiple field as constraint
			if len(pkFields) > 1 {
				primaryKey := strings.Join(utils.EscapeSqlNames(s.isPostgres, pkFields), ", ")
				indexes = append(indexes, fmt.Sprintf(s.sql.CreatePrimaryKeyStm(), tableName, primaryKey))
			}
		} else if at.Index != "" {
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

			strType, isEmbedded := s.sqlType(v.Field(j).Interface(), "")
			if isEmbedded && (at.IsEmbedded || len(at.Prefix) > 0) {
				_columns, _columnsHistory, _indexes := s.parseStruct(tableName, prefix+at.Prefix, v.Field(j).Interface())
				embedColumns = append(embedColumns, _columns...)
				embedColumnsHistory = append(embedColumnsHistory, _columnsHistory...)
				embedIndexes = append(embedIndexes, _indexes...)
				continue
			} else {
				if isEmbedded { // default type for struct is "TEXT"
					strType = s.sql.TextType()
				}
				col = append(col, strType)
			}
		}

		if at.IsNotNull {
			col = append(col, s.sql.NotNullValue())
		} else if at.IsNull {
			col = append(col, s.sql.NullValue())
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

		// primary key attribute appended after field
		if at.IsPk && len(pkFields) <= 1 {
			col = append(col, s.sql.PrimaryOption())
		}

		if s.generateComment && at.Comment == "" {
			at.Comment = createCommentFromFieldName(at.Name)
		}

		if at.Comment != "" {
			col = append(col, fmt.Sprintf(s.sql.Comment(), at.Comment))
		}

		rawCols = append(rawCols, col)
	}

	for _, f := range rawCols {
		columns = append(columns, fmt.Sprintf("  %s%s%s", utils.EscapeSqlName(s.isPostgres, f[0]), strings.Repeat(" ", maxLen-len(f[0])+1), strings.Join(f[1:], " ")))
	}

	return append(columns, embedColumns...), append(columnsHistory, embedColumnsHistory...), append(indexes, embedIndexes...)
}

func getWhenEmpty(s, s2 string) string {
	if s == "" {
		return s2
	}
	return s
}

// RemoveTable ...
func (s SqlBuilder) RemoveTable(tb interface{}) string {
	return fmt.Sprintf(s.sql.DropTableStm(), utils.EscapeSqlName(s.isPostgres, s.getTableName(tb)))
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

	if len(indexColumns) == 0 {
		indexColumns = []string{column}
	}

	return fmt.Sprintf("idx_%s", strings.Join(indexColumns, "_"))
}

// createCommentFromEnum by enum values or field name
func createCommentFromEnum(enums []string) string {
	if len(enums) > 0 {
		return fmt.Sprintf("enum values: %s", strings.Join(enums, ", "))
	}

	return ""
}

// createCommentFromFieldName by enum values or field name
func createCommentFromFieldName(column string) string {
	if _, ok := ignoredFieldComment[column]; ok {
		return ""
	}
	return strings.Replace(column, "_", " ", -1)
}

// prefix return sqlType, isEmbedded
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
func (s SqlBuilder) getTableName(t interface{}) string {
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
	if s.isTableAdds {
		name = name + "s"
	}
	return utils.ToSnakeCase(name)
}
