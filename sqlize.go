package sqlize

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	_ "github.com/pingcap/parser/test_driver"
	"github.com/sunary/sqlize/avro"
	"github.com/sunary/sqlize/mysql-builder"
	"github.com/sunary/sqlize/mysql-parser"
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
	mysqlBuilder        *mysql_builder.SqlBuilder
	mysqlParser         *mysql_parser.Parser
}

func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     "",
		migrationUpSuffix:   migrationUpSuffix,
		migrationDownSuffix: migrationDownSuffix,

		isPostgres: false,
		isLower:    false,
		sqlTag:     mysql_builder.SqlTagDefault,
	}
	for i := range opts {
		opts[i].apply(&o)
	}

	var sb *mysql_builder.SqlBuilder
	if o.isLower {
		sb = mysql_builder.NewSqlBuilder(mysql_builder.WithSqlLowercase(), mysql_builder.WithSqlTag(o.sqlTag))
	} else {
		sb = mysql_builder.NewSqlBuilder(mysql_builder.WithSqlUppercase(), mysql_builder.WithSqlTag(o.sqlTag))
	}

	return &Sqlize{
		migrationFolder:     o.migrationFolder,
		migrationUpSuffix:   o.migrationUpSuffix,
		migrationDownSuffix: o.migrationDownSuffix,
		isPostgres:          o.isPostgres,
		isLower:             o.isLower,

		mysqlBuilder: sb,
		mysqlParser:  mysql_parser.NewParser(o.isLower),
	}
}

func (s *Sqlize) FromObjects(objs ...interface{}) error {
	sqls := make([]string, len(objs))
	for i := range objs {
		sqls[i] += s.mysqlBuilder.AddTable(objs[i])
	}

	return s.FromString(strings.Join(sqls, "\n"))
}

func (s *Sqlize) FromString(sql string) error {
	return s.mysqlParser.Parser(sql)
}

func (s *Sqlize) FromMigrationFolder() error {
	sqls, err := utils.ReadPath(s.migrationFolder, s.migrationUpSuffix)
	if err != nil {
		return err
	}

	return s.FromString(strings.Join(sqls, "\n"))
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
