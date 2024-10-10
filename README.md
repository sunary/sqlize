# SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

English | [中文](README_zh.md)

SQLize is a powerful SQL toolkit for Golang, offering parsing, building, and migration capabilities.

## Features

- SQL parsing and building for multiple databases:
  - MySQL
  - PostgreSQL
  - SQLite

- SQL migration generation:
  - Create migrations from Golang models and current SQL schema
  - Generate migration versions compatible with `golang-migrate/migrate`

- Advanced functionalities:
  - Support for embedded structs
  - Avro schema generation (MySQL only)
  - Compatibility with `gorm` tags (default tag is `sql`)

## Conventions

### Default Behaviors

- Database: `mysql` (use `sql_builder.WithPostgresql()` for PostgreSQL, etc.)
- SQL syntax: Uppercase (e.g., `"SELECT * FROM user WHERE id = ?"`)
  - For lowercase, use `sql_builder.WithSqlLowercase()`
- Table naming: Singular
  - For plural (adding 's'), use `sql_builder.WithPluralTableName()`
- Comment generation: Use `sql_builder.WithCommentGenerate()`

### SQL Tag Options

- Format: Supports both `snake_case` and `camelCase` (e.g., `sql:"primary_key"` equals `sql:"primaryKey"`)
- Custom column: `sql:"column:column_name"`
- Primary key: `sql:"primary_key"`
- Foreign key: `sql:"foreign_key:user_id;references:user_id"`
- Auto increment: `sql:"auto_increment"`
- Default value: `sql:"default:CURRENT_TIMESTAMP"`
- Override datatype: `sql:"type:VARCHAR(64)"`
- Ignore field: `sql:"-"`

### Indexing

- Basic index: `sql:"index"`
- Custom index name: `sql:"index:idx_col_name"`
- Unique index: `sql:"unique"`
- Custom unique index: `sql:"unique:idx_name"`
- Composite index: `sql:"index_columns:col1,col2"` (includes unique index and primary key)
- Index type: `sql:"index_type:btree"`

### Embedded Structs

- Use `sql:"embedded"` or `sql:"squash"`
- Cannot be a pointer
- Supports prefix: `sql:"embedded_prefix:base_"`
- Fields have lowest order, except for primary key (always first)

### Data Types

- MySQL data types are implicitly changed:

```sql
	TINYINT => tinyint(4)
	INT     => int(11)
	BIGINT  => bigint(20)
```

### Important Notes

- Pointer values must be declared in the struct

### Examples

1. Using pointer values:

```golang
type sample struct {
	ID        int32 `sql:"primary_key"`
	DeletedAt *time.Time
}

now := time.Now()
newMigration.FromObjects(sample{DeletedAt: &now})
```

2. Embedded struct:

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

3. Comparing SQL schema with Go struct:

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

4. Comparing Two SQL Schemas:

```go
package main

import (
	"github.com/sunary/sqlize"
)

func main() {
	sql1 := sqlize.NewSqlize()
	sql1.FromString(`
CREATE TABLE user (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  name          VARCHAR(64),
  age           INT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_name_age ON user(name, age);
	`)

	sql2 := sqlize.NewSqlize()
	sql2.FromString(`
CREATE TABLE user (
  id            INT,
  name          VARCHAR(64),
  age           INT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME
);`)

	sql1.Diff(*sql2)
	println(sql1.StringUp())
	//ALTER TABLE `user` MODIFY COLUMN `id` int(11) AUTO_INCREMENT PRIMARY KEY;
	//ALTER TABLE `user` MODIFY COLUMN `updated_at` datetime DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP();
	//CREATE UNIQUE INDEX `idx_name_age` ON `user`(`name`, `age`);

	println(sql1.StringDown())
	//DROP INDEX `idx_name_age` ON `user`;
}
```