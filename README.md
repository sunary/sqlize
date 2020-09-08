### SQLize

Migration MySQL Compatible SQL

#### Features

+ Sql parser
+ Sql builder from objects
+ Generate sql migration from diff sql
+ Generate arvo schema


#### Examples

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

	oldMigration := sqlize.NewSqlize()
	_ = oldMigration.FromString(createStm)

	newMigration.Diff(*oldMigration)

	println(newMigration.StringUp())
	//ALTER TABLE `user` ADD COLUMN `alias` varchar(64) AFTER `id`;
	//ALTER TABLE user DROP COLUMN gender;
	//CREATE INDEX `idx_accept_tnc_at` ON `user`(`accept_tnc_at`);

	println(newMigration.StringDown())
	//ALTER TABLE user DROP COLUMN alias;
	//ALTER TABLE `user` ADD COLUMN `gender` tinyint(1) AFTER `age`;
	//DROP INDEX `idx_accept_tnc_at` ON `user`;

	println(newMigration.ArvoSchema())

	_ = newMigration.WriteFiles("demo migration")
}
```