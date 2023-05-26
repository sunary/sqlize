### SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

从golang结构和现有sql生成MySQL/PostgreSQL迁移 

#### 特征

+ Sql语法分析器
+ 对象Sql生成器
+ 根据现有sql和对象之间的差异生成“sql迁移” `sql migration`
+ 生成`arvo`架构（仅限Mysql）
+ 支持嵌入式结构体
+ 生成迁移版本-与`golang-migrate/migrate`兼容 
+ 标记选项-与“gorm”标记兼容

> **警告**: 有些函数在PostgreSQL上不起作用， 有什么问题请提issues告诉我

### 入门 

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

### 相关

* 默认情况下支持`mysql`，postgresql请使用使用选项`sql_builder.WithPostgresql（）`
* sql大写默认值，对sql小写使用选项“sql_builder.WithSqlLowercase（）”
* 支持**生成**注释，使用选项`sql_builder.WithCommentGenerate（）`
* 支持表名自动添加"s", 使用选项 `sql_builder.WithPluralTableName()`
* Accept tag convention: `snake_case` or `camelCase`, Eg: `sql:"primary_key"` equalize `sql:"primaryKey"`
* 主键/外键字段参考: `sql:"primary_key"`
* 自增: `sql:"auto_increment"`
* 给字段添加索引: `sql:"index"`
* 自定义索引名称: `sql:"index:idx_col_name"`
* 唯一索引: `sql:"unique"`
* 自定义唯一索引名称: `sql:"unique:idx_name"`
* 复合索引（包括唯一索引和主键）: `sql:"index_columns:col1,col2"`
* 索引类型 : `sql:"index_type:btree"`
* 设置默认值: `sql:"default:CURRENT_TIMESTAMP"`
* 重写数据类型: `sql:"type:VARCHAR(64)"`
* 忽略该字段,不是成到sql语句中: `sql:"-"`
* 指针值必须在结构中声明(非常重要)

```golang
type txtSample struct {
	ID        int32 `sql:"primaryKey"`
	Pid      int    `sql:"column:pid;type:int(11);default:0;NOT NULL" json:"pid"`
	sample *sample  `sql:"foreign_key:pid;references:id" json:"children"`
}
//指针值必须在结构中声明(非常重要)
// now := time.Now()
// newMigration.FromObjects(txtSample{sample:{DeletedAt: &now}})

// type sample struct {
// 	ID        int32 `sql:"primary_key"`
// 	DeletedAt *time.Time
// }

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