package sqlize

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	_ "github.com/pingcap/parser/test_driver"
	"github.com/sunary/sqlize/mysql-avro"
	"github.com/sunary/sqlize/mysql-builder"
	"github.com/sunary/sqlize/mysql-parser"
	"github.com/sunary/sqlize/utils"
)

const (
	migrationUpSuffix   = ".up.sql"
	migrationDownSuffix = ".down.sql"
	genDescription      = "--- generate by sqlize\n\n"
)

type Sqlize struct {
	migrationFolder     string
	migrationUpSuffix   string
	migrationDownSuffix string
	isLower             bool
	mysqlBuilder        *mysql_builder.SqlBuilder
	MysqlMigration      mysql_parser.Migration
}

func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     "",
		migrationUpSuffix:   migrationUpSuffix,
		migrationDownSuffix: migrationDownSuffix,

		isLower: false,
		sqlTag:  mysql_builder.SqlTagDefault,
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
		isLower:             o.isLower,

		mysqlBuilder:   sb,
		MysqlMigration: mysql_parser.NewMigration(o.isLower),
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
	return s.MysqlMigration.Parser(sql)
}

func (s *Sqlize) FromMigrationFolder() error {
	sqls, err := utils.ReadPath(s.migrationFolder, s.migrationUpSuffix)
	if err != nil {
		return err
	}

	return s.FromString(strings.Join(sqls, "\n"))
}

func (s *Sqlize) Diff(old Sqlize) {
	s.MysqlMigration.Diff(old.MysqlMigration)
}

func (s *Sqlize) StringUp() string {
	return s.MysqlMigration.MigrationUp()
}

func (s *Sqlize) StringDown() string {
	return s.MysqlMigration.MigrationDown()
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
	for i := range s.MysqlMigration.Tables {
		if len(needTables) == 0 || utils.ContainStr(needTables, s.MysqlMigration.Tables[i].Name) {
			record := mysql_avro.NewArvoSchema(s.MysqlMigration.Tables[i])
			jsonData, _ := json.Marshal(record)
			schemas = append(schemas, string(jsonData))
		}
	}

	return schemas
}
