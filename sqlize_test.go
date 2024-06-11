package sqlize

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sunary/sqlize/utils"
)

type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type addsTest struct {
	ID        int32  `sql:"primary_key;auto_increment"`
	Name      string `sql:"type:VARCHAR(64);index:name,age;unique"`
	Alias     string `sql:"-"`
	Age       int
	IsFemale  bool
	CreatedAt time.Time `sql:"default:CURRENT_TIMESTAMP"`
}

type person struct {
	ID        int32  `sql:"primary_key;auto_increment"`
	Name      string `sql:"type:VARCHAR(64);unique;index:name,age"`
	Alias     string `sql:"-"`
	Age       int
	IsFemale  bool
	CreatedAt time.Time `sql:"default:CURRENT_TIMESTAMP"`
}

type hotel struct {
	B1           Base  `sql:"embedded"`
	B2           Base  `sql:"embedded_prefix:base_"`
	ID           int32 `sql:"primary_key"`
	Name         string
	GrandOpening *time.Time
}

type city struct {
	ID     int32  `sql:"primary_key;auto_increment"`
	Name   string `sql:"column:name"`
	Region string `sql:"type:ENUM('northern','southern');default:'northern'"`
}

type movie struct {
	ID           int32  `sql:"primary_key;auto_increment"`
	Title        string `sql:"type:varchar(255)"`
	Director     string `sql:"column:director;type:varchar(255)"`
	YearReleased string `sql:"column:year_released,previous:released_at"`
}

type tpl struct {
	B1       Base   `sql:"embedded"`
	ClientID string `sql:"type:varchar(255);primary_key;index_columns:client_id,country"`
	Country  string `sql:"type:varchar(255)"`
	Email    string `sql:"type:varchar(255);unique"`
	User     *user  `sql:"foreign_key:email;references:email"`
}

func (tpl) TableName() string {
	return "three_pl"
}

type tpl_sqlite struct {
	B1       Base   `sql:"embedded"`
	ClientID string `sql:"type:text;primary_key;index_columns:client_id,country"`
	Country  string `sql:"type:text"`
	Email    string `sql:"type:text;unique"`
}

func (tpl_sqlite) TableName() string {
	return "three_pl_sqlite"
}

type user struct {
}

var (
	space = regexp.MustCompile(`\s+`)

	createAddsTestStm = `CREATE TABLE adds_tests (
		id        int(11) AUTO_INCREMENT PRIMARY KEY,
		name      varchar(64),
		age       int(11),
		is_female tinyint(1),
		created_at datetime DEFAULT CURRENT_TIMESTAMP()
	   );`
	alterAddsTestUpStm = `
	   CREATE UNIQUE INDEX idx_name_age ON adds_tests(name, age);`
	alterAddsTestDownStm = `
	   DROP INDEX idx_name_age ON adds_tests;`

	createPersonStm = `CREATE TABLE person (
 id        int(11) AUTO_INCREMENT PRIMARY KEY,
 name      varchar(64),
 age       int(11),
 is_female tinyint(1),
 created_at datetime DEFAULT CURRENT_TIMESTAMP()
);`
	alterPersonUpStm = `
CREATE UNIQUE INDEX idx_name_age ON person(name, age);`
	alterPersonDownStm = `
DROP INDEX idx_name_age ON person;`

	createHotelStm = `
CREATE TABLE hotel (
 id            int(11) PRIMARY KEY,
 name          text,
 star          tinyint(4),
 grand_opening datetime NULL,
 created_at datetime,
 updated_at datetime,
 base_created_at datetime,
 base_updated_at datetime
);`
	alterHotelUpStm = `
ALTER TABLE hotel DROP COLUMN star;`
	alterHotelDownStm = `
ALTER TABLE hotel ADD COLUMN star tinyint(4) AFTER name;`

	createCityStm = `
CREATE TABLE city (
 code   varchar(3),
 id     int(11) AUTO_INCREMENT PRIMARY KEY,
 region enum('northern','southern') DEFAULT 'northern'
);`
	alterCityUpStm = `
ALTER TABLE city DROP COLUMN code;
ALTER TABLE city ADD COLUMN name text AFTER id;`
	alterCityDownStm = `
ALTER TABLE city ADD COLUMN code varchar(3) FIRST;
ALTER TABLE city DROP COLUMN name;`

	createMovieStm = `
CREATE TABLE movie (
 id       int(11) AUTO_INCREMENT PRIMARY KEY,
 title    varchar(255),
 director varchar(255)
);`
	alterMovieUpStm = `
ALTER TABLE movie RENAME COLUMN released_at TO year_released;`
	alterMovieDownStm = `
ALTER TABLE movie RENAME COLUMN year_released TO released_at;`

	expectCreateAddsTestUp = `
CREATE TABLE adds_tests (
 id        int(11) AUTO_INCREMENT PRIMARY KEY,
 name      varchar(64),
 age       int(11),
 is_female tinyint(1),
 created_at datetime DEFAULT CURRENT_TIMESTAMP()
);
CREATE UNIQUE INDEX idx_name_age ON adds_tests(name, age);`
	expectCreateAddsTestDown = `
DROP TABLE IF EXISTS adds_tests;`

	expectCreatePersonUp = `
CREATE TABLE person (
 id        int(11) AUTO_INCREMENT PRIMARY KEY,
 name      varchar(64),
 age       int(11),
 is_female tinyint(1),
 created_at datetime DEFAULT CURRENT_TIMESTAMP()
);
CREATE UNIQUE INDEX idx_name_age ON person(name, age);`
	expectCreatePersonDown = `
DROP TABLE IF EXISTS person;`

	expectCreateHotelUp = `
CREATE TABLE hotel (
 id            int(11) PRIMARY KEY,
 name          text,
 grand_opening datetime NULL,
 created_at datetime,
 updated_at datetime,
 base_created_at datetime,
 base_updated_at datetime
);`
	expectCreateHotelDown = `
DROP TABLE IF EXISTS hotel;`

	expectCreateCityUp = `
CREATE TABLE city (
 id     int(11) AUTO_INCREMENT PRIMARY KEY,
 name   text,
 region enum('northern','southern') DEFAULT 'northern'
);`
	expectCreateCityHasCommentUp = `
CREATE TABLE city (
 id     int(11) AUTO_INCREMENT PRIMARY KEY,
 name   text COMMENT 'name',
 region enum('northern','southern') DEFAULT 'northern' COMMENT 'enum values: northern, southern'
);`
	expectCreateCityDown = `
DROP TABLE IF EXISTS city;`
	expectCreateTplUp = `
CREATE TABLE three_pl (
 client_id  varchar(255) COMMENT 'client id',
 country    varchar(255) COMMENT 'country',
 email      varchar(255) COMMENT 'email',
 created_at datetime,
 updated_at datetime
);
ALTER TABLE three_pl ADD PRIMARY KEY(client_id, country);
CREATE UNIQUE INDEX idx_email ON three_pl(email);
ALTER TABLE three_pl ADD CONSTRAINT fk_user_three_pl FOREIGN KEY (email) REFERENCES user(email);`
	expectCreateTplDown = `
DROP TABLE IF EXISTS three_pl;`
	expectCreateTplPostgresUp = `
CREATE TABLE three_pl (
 client_id VARCHAR(255),
 country   VARCHAR(255),
 email     VARCHAR(255),
 created_at TIMESTAMP,
 updated_at TIMESTAMP
);
COMMENT ON COLUMN three_pl.client_id IS 'client id';
COMMENT ON COLUMN three_pl.country IS 'country';
COMMENT ON COLUMN three_pl.email IS 'email';
ALTER TABLE three_pl ADD PRIMARY KEY(client_id, country);
CREATE UNIQUE INDEX idx_email ON three_pl(email);
ALTER TABLE three_pl ADD CONSTRAINT fk_user_three_pl FOREIGN KEY (email) REFERENCES "user"(email);`
	expectCreateTplPostgresDown = `
DROP TABLE IF EXISTS three_pl;`
	expectCreateTplSqliteUp = `
CREATE TABLE three_pl_sqlite (
 client_id  text,
 country    text,
 email      text,
 created_at text,
 updated_at text
);
ALTER TABLE three_pl_sqlite ADD PRIMARY KEY(client_id, country);
CREATE UNIQUE INDEX idx_email ON three_pl_sqlite(email);`
	expectCreateTplSqliteDown = `
DROP TABLE IF EXISTS three_pl_sqlite;`

	expectPersonArvo = `
{"type":"record","name":"person","namespace":"person","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"age","type":"int"},{"name":"is_female","type":"bool"},{"name":"created_at","type":["null",{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}]}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"person"}`
	expectHotelArvo = `
{"type":"record","name":"hotel","namespace":"hotel","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"grand_opening","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"created_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"updated_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"base_created_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"base_updated_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"hotel"}`
	expectCityArvo = `
{"type":"record","name":"city","namespace":"city","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"region","type":["null",{"connect.default":"init","connect.name":"io.debezium.data.Enum","connect.parameters":{"allowed":"northern,southern"},"connect.version":1,"type":"string"}]}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"city"}`
	expectCreateMigrationTableUp = `CREATE TABLE IF NOT EXISTS schema_migrations (
 version    bigint(20) PRIMARY KEY,
 dirty      BOOLEAN
);`
	expectCreateMigrationTableDown = `
DROP TABLE IF EXISTS schema_migrations;`
	expectMigrationVersion1Up = `
TRUNCATE schema_migrations;
INSERT INTO schema_migrations (version, dirty) VALUES (1, false);`
	expectMigrationVersion1Down = `
TRUNCATE schema_migrations;`
)

func TestSqlize_FromObjects(t *testing.T) {
	now := time.Now()

	type args struct {
		objs            []interface{}
		migrationFolder string
	}

	fromObjectMysqlTestcase := []struct {
		name              string
		generateComment   bool
		pluralTableName   bool
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{
		{
			name:            "from adds_tests object",
			pluralTableName: true,
			args: args{
				[]interface{}{addsTest{}},
				"",
			},
			wantMigrationUp:   expectCreateAddsTestUp,
			wantMigrationDown: expectCreateAddsTestDown,
			wantErr:           false,
		},
		{
			name: "from person object",
			args: args{
				[]interface{}{person{}},
				"",
			},
			wantMigrationUp:   expectCreatePersonUp,
			wantMigrationDown: expectCreatePersonDown,
			wantErr:           false,
		},
		{
			name: "from hotel object",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
				"",
			},
			wantMigrationUp:   expectCreateHotelUp,
			wantMigrationDown: expectCreateHotelDown,
			wantErr:           false,
		},
		{
			name:            "from city object",
			generateComment: true,
			args: args{
				[]interface{}{city{}},
				"/",
			},
			wantMigrationUp:   expectCreateCityHasCommentUp,
			wantMigrationDown: expectCreateCityDown,
			wantErr:           false,
		},
		{
			name:            "from tpl object",
			generateComment: true,
			args: args{
				[]interface{}{tpl{}},
				"/",
			},
			wantMigrationUp:   expectCreateTplUp,
			wantMigrationDown: expectCreateTplDown,
			wantErr:           false,
		},
		{
			name: "from all object",
			args: args{
				[]interface{}{person{}, hotel{GrandOpening: &now}, city{}},
				"/",
			},
			wantMigrationUp:   joinSql(expectCreatePersonUp, expectCreateHotelUp, expectCreateCityUp),
			wantMigrationDown: joinSql(expectCreatePersonDown, expectCreateHotelDown, expectCreateCityDown),
			wantErr:           false,
		},
	}

	for i, tt := range fromObjectMysqlTestcase {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{
				WithMigrationSuffix(".up.test", ".down.test"), WithMigrationFolder(tt.args.migrationFolder),
			}
			if tt.generateComment {
				opts = append(opts, WithCommentGenerate())
			}
			if tt.pluralTableName {
				opts = append(opts, WithPluralTableName())
			}

			if i%3 == 1 {
				opts = append(opts, WithSqlserver()) //fallback mysql
			}

			s := NewSqlize(opts...)
			if tt.args.migrationFolder == "" {
				if err := s.FromMigrationFolder(); err == nil {
					t.Errorf("FromMigrationFolder() mysql error = %v,\n wantErr = %v", err, utils.PathDoesNotExistErr)
				}
			} else if tt.args.migrationFolder == "/" {
				if err := s.FromMigrationFolder(); err != nil {
					t.Errorf("FromMigrationFolder() mysql error = %v,\n wantErr = %v", err, nil)
				}
			}

			if err := s.FromObjects(tt.args.objs...); (err != nil) != tt.wantErr {
				t.Errorf("FromObjects() mysql error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() mysql got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() mysql got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}

			if tt.args.migrationFolder == "" {
				if err := s.WriteFiles(tt.name); err != nil {
					t.Errorf("WriteFiles() mysql error = \n%v,\nwantErr = \n%v", err, nil)
				}
			} else if tt.args.migrationFolder == "/" {
				if err := s.WriteFiles(tt.name); err == nil {
					t.Errorf("WriteFiles() mysql error = \n%v,\nwantErr = \n%v", err, errors.New("read-only file system"))
				}
			}
		})
	}

	fromObjectPostgresTestcase := []struct {
		name              string
		generateComment   bool
		pluralTableName   bool
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{
		{
			name:            "from tpl object",
			generateComment: true,
			args: args{
				[]interface{}{tpl{}},
				"/",
			},
			wantMigrationUp:   expectCreateTplPostgresUp,
			wantMigrationDown: expectCreateTplPostgresDown,
			wantErr:           false,
		},
	}
	for _, tt := range fromObjectPostgresTestcase {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql(), WithCommentGenerate())
			if err := s.FromObjects(tt.args.objs...); (err != nil) != tt.wantErr {
				t.Errorf("FromObjects() postgres error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() postgres got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() postgres got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}

	fromObjectSqliteTestcase := []struct {
		name              string
		generateComment   bool
		pluralTableName   bool
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{
		{
			name:            "from tpl sqlite object",
			generateComment: true,
			args: args{
				[]interface{}{tpl_sqlite{}},
				"/",
			},
			wantMigrationUp:   expectCreateTplSqliteUp,
			wantMigrationDown: expectCreateTplSqliteDown,
			wantErr:           false,
		},
	}
	for _, tt := range fromObjectSqliteTestcase {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithSqlite())
			if err := s.FromObjects(tt.args.objs...); (err != nil) != tt.wantErr {
				t.Errorf("FromObjects() sqlite error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() sqlite got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() sqlite got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}
}

func TestSqlize_FromString(t *testing.T) {
	type args struct {
		sql string
	}

	fromStringMysqlTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{
		{
			name: "from person sql",
			args: args{
				joinSql(createPersonStm, alterPersonUpStm),
			},
			wantMigrationUp:   expectCreatePersonUp,
			wantMigrationDown: expectCreatePersonDown,
			wantErr:           false,
		},
		{
			name: "from hotel sql",
			args: args{
				joinSql(createHotelStm, alterHotelUpStm),
			},
			wantMigrationUp:   expectCreateHotelUp,
			wantMigrationDown: expectCreateHotelDown,
			wantErr:           false,
		},
		{
			name: "from city sql",
			args: args{
				joinSql(createCityStm, alterCityUpStm),
			},
			wantMigrationUp:   expectCreateCityUp,
			wantMigrationDown: expectCreateCityDown,
			wantErr:           false,
		},
	}

	for _, tt := range fromStringMysqlTestcases {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithMysql(), WithSqlUppercase())
			if err := s.FromString(tt.args.sql); (err != nil) != tt.wantErr {
				t.Errorf("FromString() mysql error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() mysql got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() mysql got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}

	fromStringPostgresTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{}
	for _, tt := range fromStringPostgresTestcases {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql())
			if err := s.FromString(tt.args.sql); (err != nil) != tt.wantErr {
				t.Errorf("FromString() postgres error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() postgres got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() postgres got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}
}

func TestSqlize_Diff(t *testing.T) {
	now := time.Now()

	type args struct {
		newObj interface{}
		oldSql string
	}

	diffMysqlTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
	}{
		{
			name: "diff person sql",
			args: args{
				person{},
				createPersonStm,
			},
			wantMigrationUp:   alterPersonUpStm,
			wantMigrationDown: alterPersonDownStm,
		},
		{
			name: "diff hotel sql",
			args: args{
				hotel{GrandOpening: &now},
				createHotelStm,
			},
			wantMigrationUp:   alterHotelUpStm,
			wantMigrationDown: alterHotelDownStm,
		},
		{
			name: "diff city sql",
			args: args{
				city{},
				createCityStm,
			},
			wantMigrationUp:   alterCityUpStm,
			wantMigrationDown: alterCityDownStm,
		},
		{
			name: "diff movie sql",
			args: args{
				movie{},
				createMovieStm,
			},
			wantMigrationUp:   alterMovieUpStm,
			wantMigrationDown: alterMovieDownStm,
		},
	}

	for _, tt := range diffMysqlTestcases {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithSqlTag("sql"), WithSqlLowercase())
			_ = s.FromObjects(tt.args.newObj)

			o := NewSqlize()
			_ = o.FromString(tt.args.oldSql)

			s.Diff(*o)
			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() mysql got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() mysql got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}

	diffPostgresTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
	}{}
	for _, tt := range diffPostgresTestcases {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql())
			_ = s.FromObjects(tt.args.newObj)

			o := NewSqlize(WithPostgresql())
			_ = o.FromString(tt.args.oldSql)

			s.Diff(*o)
			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUp() postgres got = \n%s,\nexpected = \n%s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDown() postgres got = \n%s,\nexpected = \n%s", strDown, tt.wantMigrationDown)
			}
		})
	}
}

func TestSqlize_MigrationVersion(t *testing.T) {
	now := time.Now()

	type args struct {
		models  []interface{}
		version int64
		isDirty bool
	}

	migrationVersionMysqlTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
	}{
		{
			name: "person migration version",
			args: args{
				[]interface{}{person{}},
				0,
				false,
			},
			wantMigrationUp:   expectCreatePersonUp + "\n" + expectCreateMigrationTableUp,
			wantMigrationDown: expectCreatePersonDown + "\n" + expectCreateMigrationTableDown,
		},
		{
			name: "hotel migration version",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
				1,
				false,
			},
			wantMigrationUp:   expectCreateHotelUp + "\n" + expectMigrationVersion1Up,
			wantMigrationDown: expectCreateHotelDown + "\n" + expectMigrationVersion1Down,
		},
		{
			name: "city migration version",
			args: args{
				[]interface{}{city{}},
				1,
				false,
			},
			wantMigrationUp:   expectCreateCityUp + "\n" + expectMigrationVersion1Up,
			wantMigrationDown: expectCreateCityDown + "\n" + expectMigrationVersion1Down,
		},
	}
	for _, tt := range migrationVersionMysqlTestcases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{
				WithMigrationSuffix(".up.test", ".down.test"),
				WithMigrationFolder(""),
				WithMigrationTable(utils.DefaultMigrationTable),
			}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.StringUpWithVersion(tt.args.version, tt.args.isDirty); normSql(got) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUpWithVersion() mysql got = \n%s,\nexpected = \n%s", got, tt.wantMigrationUp)
			}

			if got := s.StringDownWithVersion(tt.args.version); normSql(got) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDownWithVersion() mysql got = \n%s,\nexpected = \n%s", got, tt.wantMigrationDown)
			}

			if err := s.WriteFilesWithVersion(tt.name, tt.args.version, tt.args.isDirty); err != nil {
				t.Errorf("WriteFilesWithVersion() mysql error = \n%v,\nwantErr = \n%v", err, nil)
			}

			if err := s.WriteFilesVersion(tt.name, tt.args.version, tt.args.isDirty); err != nil {
				t.Errorf("WriteFilesVersion() mysql error = \n%v,\nwantErr = \n%v", err, nil)
			}
		})
	}

	migrationVersionPostgresTestcases := []struct {
		name              string
		args              args
		wantMigrationUp   string
		wantMigrationDown string
	}{}
	for _, tt := range migrationVersionPostgresTestcases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{
				WithMigrationSuffix(".up.test", ".down.test"),
				WithMigrationFolder(""),
				WithMigrationTable(utils.DefaultMigrationTable),
				WithPostgresql(),
			}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.StringUpWithVersion(tt.args.version, tt.args.isDirty); normSql(got) != normSql(tt.wantMigrationUp) {
				t.Errorf("StringUpWithVersion() postgres got = \n%s,\nexpected = \n%s", got, tt.wantMigrationUp)
			}

			if got := s.StringDownWithVersion(tt.args.version); normSql(got) != normSql(tt.wantMigrationDown) {
				t.Errorf("StringDownWithVersion() postgres got = \n%s,\nexpected = \n%s", got, tt.wantMigrationDown)
			}

			if err := s.WriteFilesWithVersion(tt.name, tt.args.version, tt.args.isDirty); err != nil {
				t.Errorf("WriteFilesWithVersion() postgres error = \n%v,\nwantErr = \n%v", err, nil)
			}

			if err := s.WriteFilesVersion(tt.name, tt.args.version, tt.args.isDirty); err != nil {
				t.Errorf("WriteFilesVersion() postgres error = \n%v,\nwantErr = \n%v", err, nil)
			}
		})
	}

}

func TestSqlize_HashValue(t *testing.T) {
	now := time.Now()

	type args struct {
		models []interface{}
	}

	hashValueMysqlTestcases := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "person hash value",
			args: args{
				[]interface{}{person{}},
			},
			want: -5168892191412708041,
		},
		{
			name: "hotel hash value",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
			},
			want: -3590096811374758567,
		},
		{
			name: "city hash value",
			args: args{
				[]interface{}{city{}},
			},
			want: -2026584327433441245,
		}, {
			name: "movie hash value",
			args: args{
				[]interface{}{movie{}},
			},
			want: -5515853333036032887,
		},
	}
	for _, tt := range hashValueMysqlTestcases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.HashValue(); got != tt.want {
				t.Errorf("HashValue() mysql got = \n%d,\nexpected = \n%d", got, tt.want)
			}
		})
	}

	hashValuePostgresTestcases := []struct {
		name string
		args args
		want int64
	}{}
	for _, tt := range hashValuePostgresTestcases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{WithPostgresql()}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.HashValue(); got != tt.want {
				t.Errorf("HashValue() postgres got = \n%d,\nexpected = \n%d", got, tt.want)
			}
		})
	}
}

func TestSqlize_ArvoSchema(t *testing.T) {
	now := time.Now()

	type args struct {
		models     []interface{}
		needTables []string
	}

	arvoSchemaMysqlTestcases := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "person arvo",
			args: args{
				[]interface{}{person{}},
				[]string{"person"},
			},
			want: []string{expectPersonArvo},
		},
		{
			name: "hotel arvo",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
				[]string{"hotel"},
			},
			want: []string{expectHotelArvo},
		},
		{
			name: "city arvo",
			args: args{
				[]interface{}{city{}},
				[]string{"city"},
			},
			want: []string{expectCityArvo},
		},
	}
	for _, tt := range arvoSchemaMysqlTestcases {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.ArvoSchema(tt.args.needTables...); (got != nil || tt.want != nil) && !areEqualJSON(got[0], tt.want[0]) {
				t.Errorf("ArvoSchema() mysql got = \n%v,\nexpected = \n%v", got, tt.want)
			}
		})
	}
}

func normSql(s string) string {
	s = strings.Replace(s, "`", "", -1)  // mysql escape keywords
	s = strings.Replace(s, "\"", "", -1) // postgres escape keywords
	return strings.TrimSpace(space.ReplaceAllString(s, " "))
}

func joinSql(s ...string) string {
	return strings.Join(s, "\n")
}

func areEqualJSON(s1, s2 string) bool {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}
