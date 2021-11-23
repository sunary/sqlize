package sqlize

import (
	"encoding/json"
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

type person struct {
	ID        int32  `sql:"primary_key;auto_increment"`
	Name      string `sql:"type:VARCHAR(64);unique;index:name,age"`
	Alias     string `sql:"-"`
	Age       int
	IsFemale  bool
	CreatedAt time.Time `sql:"default:CURRENT_TIMESTAMP"`
}

type hotel struct {
	B1           Base
	B2           Base  `sql:"embeddedPrefix:base_"`
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

type request struct {
	ID                       int32   `sql:"primary_key;auto_increment"`
	Sku                      string  `sql:"type:VARCHAR(64);index"`
	SiteID                   string  `sql:"type:VARCHAR(64);index:site_id"`
	CategoryID               string  `sql:"type:VARCHAR(64);index:idx_category_id"`
	FromSiteIDs              string  `sql:"type:VARCHAR(255);index:idx_from_site_ids"`
	SupplierIDs              string  `sql:"type:VARCHAR(255)"`
	AverageSale              float64 `sql:"type:DOUBLE"`
	RemainQuantity           float64 `sql:"type:DOUBLE"`
	InComingQuantity         float64 `sql:"type:DOUBLE"`
	RequestMessage           string  `sql:"type:VARCHAR(255)"`
	IsUrgent                 bool
	SuggestedQuantity        float64 `sql:"type:DOUBLE"`
	Quantity                 float64 `sql:"type:DOUBLE"`
	RequestedBy              string  `sql:"type:VARCHAR(64)"`
	Message                  string  `sql:"type:VARCHAR(255)"`
	ActionBy                 string  `sql:"type:VARCHAR(64)"`
	ReferenceID              string  `sql:"type:VARCHAR(64)"`
	ReferenceMessage         string  `sql:"type:VARCHAR(255)"`
	PurchaseReferenceID      string  `sql:"type:VARCHAR(64)"`
	PurchaseReferenceMessage string  `sql:"type:VARCHAR(255)"`
	TransferCommandID        int32
	PurchaseOrderID          int32     `sql:"index:idx_purchase_command_id"`
	CreatedAt                time.Time `sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt                time.Time `sql:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

var (
	space = regexp.MustCompile(`\s+`)

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

	requestMigration = `
CREATE TABLE request(
  id                    INT AUTO_INCREMENT PRIMARY KEY,
  sku                   VARCHAR(64),
  site_id               VARCHAR(64),
  category_id           VARCHAR(64),
  from_site_ids         VARCHAR(255),
  supplier_ids          VARCHAR(255),
  average_sale          DOUBLE,
  remain_quantity       DOUBLE,
  request_message       VARCHAR(255),
  is_urgent             BOOLEAN,
  suggested_quantity    DOUBLE,
  quantity              DOUBLE,
  requested_by          VARCHAR(64),
  message               VARCHAR(255),
  reference_id          VARCHAR(64),
  purchase_reference_id VARCHAR(64),
  response_message      VARCHAR(255),
  transfer_command_id   INT,
  purchase_order_id     INT,
  created_at            DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at            DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_sku ON request(sku);
CREATE INDEX idx_site_id ON request(site_id);
CREATE INDEX idx_category_id ON request(category_id);
CREATE INDEX idx_purchase_command_id ON request(purchase_order_id);

CREATE INDEX idx_from_site_ids ON request(from_site_ids);
ALTER TABLE request ADD COLUMN action_by VARCHAR(64) AFTER message;
ALTER TABLE request ADD COLUMN in_coming_quantity double AFTER remain_quantity;
ALTER TABLE request ADD COLUMN reference_message varchar(255) AFTER reference_id;
ALTER TABLE request ADD COLUMN purchase_reference_message varchar(255) AFTER purchase_reference_id;
ALTER TABLE request DROP COLUMN response_message;`
	expectPersonArvo = `
{"type":"record","name":"person","namespace":"person","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"age","type":"int"},{"name":"is_female","type":"bool"},{"name":"created_at","type":["null",{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}]}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"person"}`
	expectHotelArvo = `
{"type":"record","name":"hotel","namespace":"hotel","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"grand_opening","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"created_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"updated_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"base_created_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}},{"name":"base_updated_at","type":{"connect.default":"1970-01-01T00:00:00Z","connect.name":"io.debezium.time.ZonedTimestamp","connect.version":1,"type":"string"}}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"hotel"}`
	expectCityArvo = `
{"type":"record","name":"city","namespace":"city","fields":[{"name":"before","type":["null",{"type":"record","name":"Value","namespace":"","fields":[{"name":"id","type":"int"},{"name":"name","type":"string"},{"name":"region","type":["null",{"connect.default":"init","connect.name":"io.debezium.data.Enum","connect.parameters":{"allowed":"northern,southern"},"connect.version":1,"type":"string"}]}],"connect.name":""}]},{"name":"after","type":["null","Value"]},{"name":"op","type":"string"},{"name":"ts_ms","type":["null","long"]},{"name":"transaction","type":["null",{"type":"record","name":"ConnectDefault","namespace":"io.confluent.connect.avro","fields":[{"name":"id","type":"string"},{"name":"total_order","type":"long"},{"name":"data_collection_order","type":"long"}],"connect.name":""}]}],"connect.name":"city"}`
	expectMovieArvo = ``
)

func TestSqlize_FromObjects(t *testing.T) {
	now := time.Now()

	type args struct {
		objs []interface{}
	}
	tests := []struct {
		name              string
		generateComment   bool
		args              args
		wantMigrationUp   string
		wantMigrationDown string
		wantErr           bool
	}{
		{
			name: "from person object",
			args: args{
				[]interface{}{person{}},
			},
			wantMigrationUp:   expectCreatePersonUp,
			wantMigrationDown: expectCreatePersonDown,
			wantErr:           false,
		},
		{
			name: "from hotel object",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
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
			},
			wantMigrationUp:   expectCreateCityHasCommentUp,
			wantMigrationDown: expectCreateCityDown,
			wantErr:           false,
		},
		{
			name: "from all object",
			args: args{
				[]interface{}{person{}, hotel{GrandOpening: &now}, city{}},
			},
			wantMigrationUp:   joinSql(expectCreatePersonUp, expectCreateHotelUp, expectCreateCityUp),
			wantMigrationDown: joinSql(expectCreatePersonDown, expectCreateHotelDown, expectCreateCityDown),
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{
				WithMigrationSuffix(".up.test", ".down.test"), WithMigrationFolder(""),
			}
			if tt.generateComment {
				opts = append(opts, WithCommentGenerate())
			}

			s := NewSqlize(opts...)
			if err := s.FromMigrationFolder(); err == nil {
				t.Errorf("Mysql FromMigrationFolder() error = %v,\n wantErr = %v", nil, utils.PathDoesNotExistErr)
			}

			if err := s.FromObjects(tt.args.objs...); (err != nil) != tt.wantErr {
				t.Errorf("Mysql FromObjects() error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Mysql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Mysql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
			}

			if err := s.WriteFiles("demo"); err != nil {
				t.Errorf("Mysql WriteFiles() error = %v,\n wantErr = %v", err, nil)
			}
		})
	}

	t.Skip()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql())
			if err := s.FromObjects(tt.args.objs...); (err != nil) != tt.wantErr {
				t.Errorf("Postgresql FromObjects() error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Postgresql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Postgresql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
			}
		})
	}
}

func TestSqlize_FromString(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithSqlUppercase())
			if err := s.FromString(tt.args.sql); (err != nil) != tt.wantErr {
				t.Errorf("Mysql FromString() error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Mysql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Mysql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
			}
		})
	}

	t.Skip()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql())
			if err := s.FromString(tt.args.sql); (err != nil) != tt.wantErr {
				t.Errorf("Postgresql FromString() error = %v,\n wantErr = %v", err, tt.wantErr)
			}

			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Postgresql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Postgresql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
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
	tests := []struct {
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
		{
			name: "diff request sql",
			args: args{
				request{},
				requestMigration,
			},
			wantMigrationUp:   "",
			wantMigrationDown: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithSqlLowercase(), WithSqlTag("sql"))
			_ = s.FromObjects(tt.args.newObj)

			o := NewSqlize()
			_ = o.FromString(tt.args.oldSql)

			s.Diff(*o)
			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Mysql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Mysql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
			}
		})
	}

	t.Skip()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSqlize(WithPostgresql())
			_ = s.FromObjects(tt.args.newObj)

			o := NewSqlize()
			_ = o.FromString(tt.args.oldSql)

			s.Diff(*o)
			if strUp := s.StringUp(); normSql(strUp) != normSql(tt.wantMigrationUp) {
				t.Errorf("Postgresql StringUp() got = %s,\n expected = %s", strUp, tt.wantMigrationUp)
			}

			if strDown := s.StringDown(); normSql(strDown) != normSql(tt.wantMigrationDown) {
				t.Errorf("Postgresql StringDown() got = %s,\n expected = %s", strDown, tt.wantMigrationDown)
			}
		})
	}
}

func TestSqlize_ArvoSchema(t *testing.T) {
	now := time.Now()

	type args struct {
		models       []interface{}
		needTables   []string
		isPostgresql bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "person arvo",
			args: args{
				[]interface{}{person{}},
				[]string{"person"},
				false,
			},
			want: []string{expectPersonArvo},
		},
		{
			name: "hotel arvo",
			args: args{
				[]interface{}{hotel{GrandOpening: &now}},
				[]string{"hotel"},
				false,
			},
			want: []string{expectHotelArvo},
		},
		{
			name: "city arvo",
			args: args{
				[]interface{}{city{}},
				[]string{"city"},
				false,
			},
			want: []string{expectCityArvo},
		}, {
			name: "movie arvo",
			args: args{
				[]interface{}{movie{}},
				[]string{"movie"},
				true,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []SqlizeOption{}
			if tt.args.isPostgresql {
				opts = append(opts, WithPostgresql())
			}
			s := NewSqlize(opts...)

			s.FromObjects(tt.args.models...)
			if got := s.ArvoSchema(tt.args.needTables...); (got != nil || tt.want != nil) && !areEqualJSON(got[0], tt.want[0]) {
				t.Errorf("ArvoSchema() got = %v,\n expected = %v", got, tt.want)
			}
		})
	}
}

func normSql(s string) string {
	s = strings.Replace(s, "`", "", -1)
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
