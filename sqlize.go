package sqlize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/pingcap/parser/test_driver" // driver parser
	"github.com/sunary/sqlize/element"
	"github.com/sunary/sqlize/export/avro"
	"github.com/sunary/sqlize/export/mermaidjs"
	sql_builder "github.com/sunary/sqlize/sql-builder"
	sql_parser "github.com/sunary/sqlize/sql-parser"
	sql_templates "github.com/sunary/sqlize/sql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	genDescription = "/* generate by sqlize */\n\n"
	emptyMigration = "/* empty */"
)

// Sqlize ...
type Sqlize struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	migrationTable      string
	dialect             sql_templates.SqlDialect
	lowercase           bool
	pluralTableName     bool
	sqlBuilder          *sql_builder.SqlBuilder
	parser              *sql_parser.Parser
}

// NewSqlize ...
func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     utils.DefaultMigrationFolder,
		migrationUpSuffix:   utils.DefaultMigrationUpSuffix,
		migrationDownSuffix: utils.DefaultMigrationDownSuffix,
		migrationTable:      utils.DefaultMigrationTable,

		dialect:          sql_templates.MysqlDialect,
		lowercase:        false,
		sqlTag:           sql_builder.SqlTagDefault,
		pluralTableName:  false,
		generateComment:  false,
		ignoreFieldOrder: false,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	opt := []sql_builder.SqlBuilderOption{sql_builder.WithSqlTag(o.sqlTag), sql_builder.WithDialect(o.dialect)}

	if o.lowercase {
		opt = append(opt, sql_builder.WithSqlLowercase())
	}

	if o.generateComment {
		opt = append(opt, sql_builder.WithCommentGenerate())
	}

	if o.pluralTableName {
		opt = append(opt, sql_builder.WithPluralTableName())
	}

	sb := sql_builder.NewSqlBuilder(opt...)

	return &Sqlize{
		migrationFolder:     o.migrationFolder,
		migrationUpSuffix:   o.migrationUpSuffix,
		migrationDownSuffix: o.migrationDownSuffix,
		migrationTable:      o.migrationTable,
		dialect:             o.dialect,
		lowercase:           o.lowercase,
		pluralTableName:     o.pluralTableName,
		sqlBuilder:          sb,
		parser:              sql_parser.NewParser(o.dialect, o.lowercase, o.ignoreFieldOrder),
	}
}

// FromObjects load from objects
func (s *Sqlize) FromObjects(objs ...interface{}) error {
	m := map[string]string{}
	for i := range objs {
		ob, tb := s.sqlBuilder.GetTableName(objs[i])
		m[ob] = tb
	}

	s.sqlBuilder.MappingTables(m)

	for i := range objs {
		if err := s.FromString(s.sqlBuilder.AddTable(objs[i])); err != nil {
			return err
		}
	}

	return nil
}

// FromString load migration from sql
func (s *Sqlize) FromString(sql string) error {
	return s.parser.Parser(sql)
}

// FromMigrationFolder load migration from folder `migrations`
func (s *Sqlize) FromMigrationFolder() error {
	sqls, err := utils.ReadPath(s.migrationFolder, s.migrationUpSuffix)
	if err != nil {
		return err
	}

	for _, sql := range sqls {
		if err := s.FromString(sql); err != nil {
			return err
		}
	}

	return nil
}

// HashValue ...
func (s Sqlize) HashValue() int64 {
	return s.parser.HashValue()
}

// Diff differ between 2 migrations
func (s Sqlize) Diff(old Sqlize) {
	if s.dialect != old.dialect {
		panic("unable to differentiate between two distinct dialects.")
	}

	s.parser.Diff(*old.parser)
}

// StringUp migration up
func (s Sqlize) StringUp() string {
	return s.parser.MigrationUp()
}

// StringUpWithVersion migration up with version
func (s Sqlize) StringUpWithVersion(ver int64, dirty bool) string {
	return s.StringUp() + "\n" + s.migrationUpVersion(ver, dirty)
}

// StringDown migration down
func (s Sqlize) StringDown() string {
	return s.parser.MigrationDown()
}

// StringDownWithVersion migration down with version
func (s Sqlize) StringDownWithVersion(ver int64) string {
	return s.StringDown() + "\n" + s.migrationDownVersion(ver)
}

func (s Sqlize) writeFiles(name, migUp, migDown string) error {
	if migUp == "" && migDown == "" {
		return nil
	}

	if migUp == "" {
		migUp = emptyMigration
	}

	if migDown == "" {
		migDown = emptyMigration
	}

	fileName := utils.MigrationFileName(name)

	filePath := filepath.Join(s.migrationFolder, fileName+s.migrationUpSuffix)
	err := os.WriteFile(filePath, []byte(genDescription+migUp), 0644)
	if err != nil {
		return err
	}

	if s.migrationDownSuffix != "" && s.migrationDownSuffix != s.migrationUpSuffix {
		filePath := filepath.Join(s.migrationFolder, fileName+s.migrationDownSuffix)
		err := os.WriteFile(filePath, []byte(genDescription+migDown), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteFiles create migration files
func (s Sqlize) WriteFiles(name string) error {
	return s.writeFiles(name,
		s.StringUp(),
		s.StringDown())
}

// WriteFilesVersion create migration version only
func (s Sqlize) WriteFilesVersion(name string, ver int64, dirty bool) error {
	return s.writeFiles(name,
		s.migrationUpVersion(ver, dirty),
		s.migrationDownVersion(ver))
}

// WriteFilesWithVersion create migration files with version
func (s Sqlize) WriteFilesWithVersion(name string, ver int64, dirty bool) error {
	return s.writeFiles(name,
		s.StringUp()+"\n\n"+s.migrationUpVersion(ver, dirty),
		s.StringDown()+"\n\n"+s.migrationDownVersion(ver))
}

func (s Sqlize) migrationUpVersion(ver int64, dirty bool) string {
	tmp := sql_templates.NewSql(s.dialect, s.lowercase)
	if ver == 0 {
		return fmt.Sprintf(tmp.CreateTableMigration(), s.migrationTable)
	}

	return fmt.Sprintf(tmp.InsertMigrationVersion(), s.migrationTable, s.migrationTable, ver, dirty)
}

func (s Sqlize) migrationDownVersion(ver int64) string {
	tmp := sql_templates.NewSql(s.dialect, s.lowercase)
	if ver == 0 {
		return fmt.Sprintf(tmp.DropTableMigration(), s.migrationTable)
	}

	return fmt.Sprintf(tmp.RollbackMigrationVersion(), s.migrationTable)
}

func (s Sqlize) selectTable(needTables ...string) []element.Table {
	tables := make([]element.Table, 0, len(needTables))

	for i := range s.parser.Migration.Tables {
		if len(needTables) == 0 || utils.ContainStr(needTables, s.parser.Migration.Tables[i].Name) {
			tables = append(tables, s.parser.Migration.Tables[i])
		}
	}

	return tables
}

// MermaidJsErd export MermaidJs ERD
func (s Sqlize) MermaidJsErd(needTables ...string) string {
	mm := mermaidjs.NewMermaidJs(s.selectTable(needTables...))
	return mm.String()
}

// MermaidJsLive export MermaidJs Live
func (s Sqlize) MermaidJsLive(needTables ...string) string {
	mm := mermaidjs.NewMermaidJs(s.selectTable(needTables...))
	return mm.Live()
}

// ArvoSchema export arvo schema, support mysql only
func (s Sqlize) ArvoSchema(needTables ...string) []string {
	if s.dialect != sql_templates.MysqlDialect {
		return nil
	}

	tables := s.selectTable(needTables...)
	schemas := make([]string, 0, len(tables))
	for i := range tables {
		record := avro.NewArvoSchema(tables[i])
		jsonData, _ := json.Marshal(record)
		schemas = append(schemas, string(jsonData))
	}

	return schemas
}
