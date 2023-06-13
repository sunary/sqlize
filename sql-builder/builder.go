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
	compileKeepEnumChar = regexp.MustCompile("[^0-9A-Za-z,\\-_]+")
)

// SqlBuilder ...
type SqlBuilder struct {
	sql             *sql_templates.Sql
	isPostgres      bool
	sqlTag          string
	generateComment bool
	pluralTableName bool
}

// NewSqlBuilder ...
func NewSqlBuilder(opts ...SqlBuilderOption) *SqlBuilder {
	o := sqlBuilderOptions{
		isLower:         false,
		isPostgres:      false,
		pluralTableName: false,
		sqlTag:          SqlTagDefault,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	return &SqlBuilder{
		sql:             sql_templates.NewSql(o.isPostgres, o.isLower),
		isPostgres:      o.isPostgres,
		sqlTag:          o.sqlTag,
		pluralTableName: o.pluralTableName,
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
		stag := field.Tag.Get(s.sqlTag)
		if stag == "-" {
			continue
		}

		at := attrs{
			Name: prefix + utils.ToSnakeCase(field.Name),
		}

		xstag := strings.Split(stag, ";")
		for _, ot := range xstag {
			hasBreak := false

			// normalize tag, convert `primaryKey` => `primary_key`
			normTag := utils.ToSnakeCase(ot)

			switch {
			case strings.HasPrefix(normTag, prefixColumn):
				columnNames := strings.Split(trimPrefix(ot, prefixColumn), prefixPreviousName)
				if len(columnNames) == 1 {
					at.Name = columnNames[0]
				} else {
					columnsHistory = append(columnsHistory, [2]string{columnNames[1], columnNames[0]})
					at.Name = columnNames[1]
				}
				at.Name = prefix + at.Name

			case strings.HasPrefix(normTag, prefixEmbedded):
				at.Prefix = trimPrefix(ot, prefixEmbedded)
				at.IsEmbedded = true

			case strings.HasPrefix(normTag, prefixForeignKey) || strings.HasPrefix(normTag, prefixAssociationFk):
				at.IsFk = true
				hasBreak = true

			case strings.HasPrefix(normTag, prefixType):
				at.Type = trimPrefix(ot, prefixType)
				if s.generateComment && at.Comment == "" && strings.HasPrefix(strings.ToLower(at.Type), tagEnum) {
					enumRaw := ot[len(prefixType)+len(tagEnum):] // this tag is safe, remove prefix: "type:ENUM"
					enumStr := compileKeepEnumChar.ReplaceAllString(enumRaw, "")
					at.Comment = createCommentFromEnum(strings.Split(enumStr, ","))
				}

			case strings.HasPrefix(normTag, prefixDefault):
				at.Value = fmt.Sprintf(s.sql.DefaultOption(), trimPrefix(ot, (prefixDefault)))

			case strings.HasPrefix(normTag, prefixComment):
				at.Comment = trimPrefix(ot, prefixComment)

			case normTag == tagIsPrimaryKey:
				at.IsPk = true

			case normTag == tagIsIndex:
				at.Index = getWhenEmpty(at.Index, createIndexName("", nil, at.Name))
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case normTag == tagIsUniqueIndex:
				at.IsUnique = true
				at.Index = getWhenEmpty(at.Index, createIndexName("", nil, at.Name))
				if at.IndexColumns == "" {
					at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)
				}

			case strings.HasPrefix(normTag, prefixIndex):
				idxFields := strings.Split(trimPrefix(ot, prefixIndex), ",")
				at.Index = createIndexName(prefix, idxFields, at.Name)

				if len(idxFields) > 1 {
					at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")
				} else {
					at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)
				}

			case strings.HasPrefix(normTag, prefixUniqueIndex):
				at.IsUnique = true
				at.Index = createIndexName(prefix, []string{trimPrefix(ot, prefixUniqueIndex)}, at.Name)
				at.IndexColumns = utils.EscapeSqlName(s.isPostgres, at.Name)

			case strings.HasPrefix(normTag, prefixIndexColumns):
				idxFields := strings.Split(trimPrefix(ot, prefixIndexColumns), ",")
				if at.IsPk {
					pkFields = idxFields
				}

				at.Index = createIndexName(prefix, idxFields, at.Name)
				at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, idxFields), ", ")

			case strings.HasPrefix(normTag, prefixIndexType):
				at.Index = getWhenEmpty(at.Index, createIndexName(prefix, nil, at.Name))

				if len(at.IndexColumns) == 0 {
					at.IndexColumns = strings.Join(utils.EscapeSqlNames(s.isPostgres, []string{at.Name}), ", ")
				}
				at.IndexType = trimPrefix(ot, prefixIndexType)

			case normTag == tagIsNull:
				at.IsNotNull = true

			case normTag == tagIsNotNull:
				at.IsNotNull = true

			case normTag == tagIsAutoIncrement:
				at.IsAutoIncr = true

			case normTag == tagIsSquash, normTag == tagIsEmbedded:
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

// because tag is norm
func trimPrefix(ot, prefix string) string {
	if strings.HasPrefix(ot, prefix) {
		return ot[len(prefix):]
	}

	// all prefix tag is end with `:`
	return ot[strings.Index(ot, ":")+1:]
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

// getTableName read from TableNameFn or parse table name from model as snake_case
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
	if s.pluralTableName {
		name = name + "s"
	}
	return utils.ToSnakeCase(name)
}
