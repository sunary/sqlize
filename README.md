### SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

Generate MySQL/PostgreSQL Migration from golang struct and existing sql

#### Features

+ Sql parser
+ Sql builder from objects
+ Generate `sql migration` from diff between existed sql and objects
+ Generate `arvo` schema (Mysql only)
+ Support embedded struct
+ Generate migration version - compatible with `golang-migrate/migrate`
+ Tag options - compatible with `gorm` tag

> **WARNING**: some functions doesn't work on PostgreSQL, let me know of any issues

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
	Name        string `sql:"type:VARCHAR(64);unique;index:name,age"`
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

* `mysql` by default, using option `sql_builder.WithPostgresql()` for `postgresql`
* sql uppercase default, using option `sql_builder.WithSqlLowercase()` for sql lowercase
* support **generate** comment, using option `sql_builder.WithCommentGenerate()`
* primary key: `sql:"primary_key"`
* auto increment: `sql:"auto_increment"`
* index on a single column: `sql:"index"` or `sql:"index:col_name"`, index name will be `idx_col_name`
* index on a single column (custom name): `sql:"index:idx_name"`
* composite index (can not custom): `sql:"index:col1,col2"`, index name will be `idx_col1_col2`
* index type: `sql:"index_type:btree"`
* unique: `sql:"unique"`
* set default value: `sql:"default:CURRENT_TIMESTAMP"`
* override datatype: `sql:"type:VARCHAR(64)"`
* ignore: `sql:"-"`
* pointer value must be declare in struct

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
* an embedded struct can not be pointer, also support prefix: `sql:"embedded_prefix:base_"`

```golang
type Base struct {
	ID        int32 `sql:"primary_key"`
	CreatedAt time.Time
}
type sample struct {
	Base
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