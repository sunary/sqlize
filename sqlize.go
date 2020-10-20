package sqlize

import (
	"encoding/json"
	"io/ioutil"

	_ "github.com/pingcap/parser/test_driver"
	"github.com/sunary/sqlize/avro"
	"github.com/sunary/sqlize/mysql-parser"
	"github.com/sunary/sqlize/sql-builder"
	"github.com/sunary/sqlize/utils"
)

const (
	migrationUpSuffix   = ".up.sql"
	migrationDownSuffix = ".down.sql"
	genDescription      = "# generate by sqlize\n\n"
)

type Sqlize struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	isPostgres          bool
	isLower             bool
	sqlBuilder          *sql_builder.SqlBuilder
	mysqlParser         *mysql_parser.Parser
}

func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     "",
		migrationUpSuffix:   migrationUpSuffix,
		migrationDownSuffix: migrationDownSuffix,

		isPostgres: false,
		isLower:    false,
		sqlTag:     sql_builder.SqlTagDefault,
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
	sb := sql_builder.NewSqlBuilder(opt...)

	return &Sqlize{
		migrationFolder:     o.migrationFolder,
		migrationUpSuffix:   o.migrationUpSuffix,
		migrationDownSuffix: o.migrationDownSuffix,
		isPostgres:          o.isPostgres,
		isLower:             o.isLower,

		sqlBuilder:  sb,
		mysqlParser: mysql_parser.NewParser(o.isLower),
	}
}

func (s *Sqlize) FromObjects(objs ...interface{}) error {
	for i := range objs {
		if err := s.FromString(s.sqlBuilder.AddTable(objs[i])); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sqlize) FromString(sql string) error {
	return s.mysqlParser.Parser(sql)
}

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

func (s *Sqlize) Diff(old Sqlize) {
	s.mysqlParser.Diff(*old.mysqlParser)
}

func (s *Sqlize) StringUp() string {
	return s.mysqlParser.MigrationUp()
}

func (s *Sqlize) StringDown() string {
	return s.mysqlParser.MigrationDown()
}

func (s *Sqlize) WriteFiles(name string) error {
	migrationUp := s.StringUp()
	if len(migrationUp) > 0 {
		migrationName := utils.MigrationFileName(name)

		err := ioutil.WriteFile(s.migrationFolder+migrationName+s.migrationUpSuffix, []byte(genDescription+migrationUp), 0644)
		if err != nil {
			return err
		}

		if s.migrationDownSuffix != "" && s.migrationDownSuffix != s.migrationUpSuffix {
			err := ioutil.WriteFile(s.migrationFolder+migrationName+s.migrationDownSuffix, []byte(genDescription+s.StringDown()), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s Sqlize) ArvoSchema(needTables ...string) []string {
	schemas := make([]string, 0)
	for i := range s.mysqlParser.Migration.Tables {
		if len(needTables) == 0 || utils.ContainStr(needTables, s.mysqlParser.Migration.Tables[i].Name) {
			record := avro.NewArvoSchema(s.mysqlParser.Migration.Tables[i])
			jsonData, _ := json.Marshal(record)
			schemas = append(schemas, string(jsonData))
		}
	}

	return schemas
}
