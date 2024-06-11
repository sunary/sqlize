### SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

English | [中文](README_zh.md)

Generate SQL migration schema from Golang models and the current SQL schema, support:

- [x] MySQL
- [x] Postgres
- [x] Sqlite
- [ ] Sql Server

#### Features

+ Sql parser (MySQL, Postgres, Sqlite)
+ Sql builder from objects (MySQL, Postgres, Sqlite)
+ Generate `sql migration` from Golang models and the current SQL schema
+ Generate `arvo` schema (Mysql only)
+ Support embedded struct
+ Generate migration version - compatible with `golang-migrate/migrate`
+ Tag options - compatible with `gorm` tag (default tag is `sql`)


### Getting Started

```go
package main

import (
	"time"
	
	"github.com/sunary/sqlize"
)

type user struct {
	ID          int32  `sql:"primary_key;auto_increment"`
	Alias       string `sql:"type:VARCHAR(64)"`
	Name        string `sql:"type:VARCHAR(64);unique;index_columns:name,age"`
	Age         int
	Bio         string
	IgnoreMe    string     `sql:"-"`
	AcceptTncAt *time.Time `sql:"index:idx_accept_tnc_at"`
	CreatedAt   time.Time  `sql:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `sql:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;index:idx_updated_at"`
}

func (user) TableName() string {
	return "user"
}

var createStm = `
CREATE TABLE user (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(64),
  age           INT,
  bio           TEXT,
  gender        BOOL,
  accept_tnc_at DATETIME NULL,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_name_age ON user(name, age);
CREATE INDEX idx_updated_at ON user(updated_at);`

func main() {
	n := time.Now()
	newMigration := sqlize.NewSqlize(sqlize.WithSqlTag("sql"), sqlize.WithMigrationFolder(""))
	_ = newMigration.FromObjects(user{AcceptTncAt: &n})

	println(newMigration.StringUp())
	//CREATE TABLE `user` (
	//	`id`            int(11) AUTO_INCREMENT PRIMARY KEY,
	//	`alias`         varchar(64),
	//	`name`          varchar(64),
	//	`age`           int(11),
	//	`bio`           text,
	//	`accept_tnc_at` datetime NULL,
	//	`created_at`    datetime DEFAULT CURRENT_TIMESTAMP(),
	//	`updated_at`    datetime DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP()
	//);
	//CREATE UNIQUE INDEX `idx_name_age` ON `user`(`name`, `age`);
	//CREATE INDEX `idx_accept_tnc_at` ON `user`(`accept_tnc_at`);
	//CREATE INDEX `idx_updated_at` ON `user`(`updated_at`);

	println(newMigration.StringDown())
	//DROP TABLE IF EXISTS `user`;

	oldMigration := sqlize.NewSqlize(sqlize.WithMigrationFolder(""))
	//_ = oldMigration.FromMigrationFolder()
	_ = oldMigration.FromString(createStm)

	newMigration.Diff(*oldMigration)

	println(newMigration.StringUp())
	//ALTER TABLE `user` ADD COLUMN `alias` varchar(64) AFTER `id`;
	//ALTER TABLE `user` DROP COLUMN `gender`;
	//CREATE INDEX `idx_accept_tnc_at` ON `user`(`accept_tnc_at`);

	println(newMigration.StringDown())
	//ALTER TABLE `user` DROP COLUMN `alias`;
	//ALTER TABLE `user` ADD COLUMN `gender` tinyint(1) AFTER `age`;
	//DROP INDEX `idx_accept_tnc_at` ON `user`;

	println(newMigration.ArvoSchema())
	//...

	_ = newMigration.WriteFiles("demo migration")
}
```

### Convention

* `mysql` by default, using options like `sql_builder.WithPostgresql()` for `postgresql`, ...
* Sql syntax uppercase (Eg: `"SELECT * FROM user WHERE id = ?"`) default, using option `sql_builder.WithSqlLowercase()` for lowercase
* Support **generate** comment, using option `sql_builder.WithCommentGenerate()`
* Support automatic addition of `s` to table names (plural naming convention), using option `sql_builder.WithPluralTableName()`
* Accept tag convention: `snake_case` or `camelCase`, Eg: `sql:"primary_key"` equalize `sql:"primaryKey"`
* Custom column name: `sql:"column:column_name"`
* Primary key for this field: `sql:"primary_key"`
* Foreign key: `sql:"foreign_key:user_id;references:user_id"`
* Auto increment: `sql:"auto_increment"`
* Indexing this field: `sql:"index"`
* Custom index name: `sql:"index:idx_col_name"`
* Unique indexing this field: `sql:"unique"`
* Custome unique index name: `sql:"unique:idx_name"`
* Composite index (include unique index and primary key): `sql:"index_columns:col1,col2"`
* Index type: `sql:"index_type:btree"`
* Set default value: `sql:"default:CURRENT_TIMESTAMP"`
* Override datatype: `sql:"type:VARCHAR(64)"`
* Ignore: `sql:"-"`
* Pointer value must be declare in struct

```golang
type sample struct {
	ID        int32 `sql:"primary_key"`
	DeletedAt *time.Time
}

now := time.Now()
newMigration.FromObjects(sample{DeletedAt: &now})
```

* `mysql` data type will be changed implicitly:

```sql
TINYINT => tinyint(4)
INT     => int(11)
BIGINT  => bigint(20)
```

* fields belong to embedded struct have the lowest order, except `primary key` always first
* an embedded struct (`sql:"embedded"` or `sql:"squash"`) can not be pointer, also support prefix: `sql:"embedded_prefix:base_"`

```golang
type Base struct {
	ID        int32 `sql:"primary_key"`
	CreatedAt time.Time
}
type sample struct {
	Base `sql:"embedded"`
	User string
}

newMigration.FromObjects(sample{})

/*
CREATE TABLE sample (
 id         int(11) PRIMARY KEY,
 user       text,
 created_at datetime
);
*/
```
