package sqlize

import (
	"encoding/json"
	"io/ioutil"

	_ "github.com/pingcap/parser/test_driver" // driver parser
	"github.com/sunary/sqlize/avro"
	"github.com/sunary/sqlize/sql-builder"
	"github.com/sunary/sqlize/sql-parser"
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
	isPostgres          bool
	isLower             bool
	sqlBuilder          *sql_builder.SqlBuilder
	parser              *sql_parser.Parser
}

// NewSqlize ...
func NewSqlize(opts ...SqlizeOption) *Sqlize {
	o := sqlizeOptions{
		migrationFolder:     "",
		migrationUpSuffix:   utils.MigrationUpSuffix,
		migrationDownSuffix: utils.MigrationDownSuffix,

		isPostgres: false,
		isLower:    false,
		sqlTag:     sql_builder.SqlTagDefault,
		hasComment: false,
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

	if o.hasComment {
		opt = append(opt, sql_builder.HasComment())
	}
	sb := sql_builder.NewSqlBuilder(opt...)

	return &Sqlize{
		migrationFolder:     o.migrationFolder,
		migrationUpSuffix:   o.migrationUpSuffix,
		migrationDownSuffix: o.migrationDownSuffix,
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

// Diff differ between 2 migrations
func (s *Sqlize) Diff(old Sqlize) {
	s.parser.Diff(*old.parser)
}

// StringUp migration up
func (s *Sqlize) StringUp() string {
	return s.parser.MigrationUp()
}

// StringDown migration down
func (s *Sqlize) StringDown() string {
	return s.parser.MigrationDown()
}

// WriteFiles create migration file
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
