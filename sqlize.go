package sqlize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	_ "github.com/pingcap/parser/test_driver" // driver parser
	"github.com/sunary/sqlize/avro"
	"github.com/sunary/sqlize/sql-builder"
	"github.com/sunary/sqlize/sql-parser"
	sql_templates "github.com/sunary/sqlize/sql-templates"
	"github.com/sunary/sqlize/utils"
)

const (
	genDescription = "# generate by sqlize\n\n"
)

// Sqlize ...
type Sqlize struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	isMigrationCheck    bool
	migrationTable      string
	isPostgres          bool
	isLower             bool
	sqlBuilder          *sql_builder.SqlBuilder
	parser              *sql_parser.Parser
}

// NewSqlize ...
func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     utils.DefaultMigrationFolder,
		migrationUpSuffix:   utils.DefaultMigrationUpSuffix,
		migrationDownSuffix: utils.DefaultMigrationDownSuffix,
		isMigrationCheck:    false,
		migrationTable:      utils.DefaultMigrationTable,

		isPostgres:      false,
		isLower:         false,
		sqlTag:          sql_builder.SqlTagDefault,
		generateComment: false,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	opt := []sql_builder.SqlBuilderOption{sql_builder.WithSqlTag(o.sqlTag)}
	if o.isPostgres {
		opt = append(opt, sql_builder.WithPostgresql())
	}

	if o.isLower {
		opt = append(opt, sql_builder.WithSqlLowercase())
	}

	if o.generateComment {
		opt = append(opt, sql_builder.WithCommentGenerate())
	}
	sb := sql_builder.NewSqlBuilder(opt...)

	return &Sqlize{
		migrationFolder:     o.migrationFolder,
		migrationUpSuffix:   o.migrationUpSuffix,
		migrationDownSuffix: o.migrationDownSuffix,
		isMigrationCheck:    o.isMigrationCheck,
		migrationTable:      o.migrationTable,
		isPostgres:          o.isPostgres,
		isLower:             o.isLower,

		sqlBuilder: sb,
		parser:     sql_parser.NewParser(o.isPostgres, o.isLower),
	}
}

// FromObjects load from objects
func (s *Sqlize) FromObjects(objs ...interface{}) error {
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

func (s Sqlize) writeFiles(name string, ver int64, dirty bool) error {
	migrationUp := s.StringUp()
	if len(migrationUp) > 0 {
		migrationName := utils.MigrationFileName(name)

		if ver >= 0 {
			migrationUp = genDescription + migrationUp + "\n" + s.migrationUpVersion(ver, dirty)
		} else {
			migrationUp = genDescription + migrationUp
		}
		err := ioutil.WriteFile(s.migrationFolder+migrationName+s.migrationUpSuffix, []byte(migrationUp), 0644)
		if err != nil {
			return err
		}

		migrationDown := genDescription + s.StringDown()
		if ver >= 0 {
			migrationDown = migrationDown + "\n" + s.migrationDownVersion(ver)
		}
		if s.migrationDownSuffix != "" && s.migrationDownSuffix != s.migrationUpSuffix {
			err := ioutil.WriteFile(s.migrationFolder+migrationName+s.migrationDownSuffix, []byte(migrationDown), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteFiles create migration files
func (s Sqlize) WriteFiles(name string) error {
	return s.writeFiles(name, -1, false)
}

// WriteFilesWithVersion create migration files with version
func (s Sqlize) WriteFilesWithVersion(name string, ver int64, dirty bool) error {
	return s.writeFiles(name, ver, dirty)
}

func (s Sqlize) migrationUpVersion(ver int64, dirty bool) string {
	tmp := sql_templates.NewSql(s.isPostgres, s.isLower)
	if ver == 0 {
		return fmt.Sprintf(tmp.CreateTableMigration(), s.migrationTable)
	}

	return fmt.Sprintf(tmp.InsertMigrationVersion(), s.migrationTable, s.migrationTable, ver, dirty)
}

func (s Sqlize) migrationDownVersion(ver int64) string {
	tmp := sql_templates.NewSql(s.isPostgres, s.isLower)
	if ver == 0 {
		return fmt.Sprintf(tmp.DropTableMigration(), s.migrationTable)
	}

	return fmt.Sprintf(tmp.RollbackMigrationVersion(), s.migrationTable)
}

// ArvoSchema export arvo schema
func (s Sqlize) ArvoSchema(needTables ...string) []string {
	if s.isPostgres {
		return nil
	}

	schemas := make([]string, 0)
	for i := range s.parser.Migration.Tables {
		if len(needTables) == 0 || utils.ContainStr(needTables, s.parser.Migration.Tables[i].Name) {
			record := avro.NewArvoSchema(s.parser.Migration.Tables[i])
			jsonData, _ := json.Marshal(record)
			schemas = append(schemas, string(jsonData))
		}
	}

	return schemas
}
