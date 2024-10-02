### SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

[English](README.md) | 中文

SQLize 是一个强大的 Golang SQL 工具包，提供解析、构建和迁移功能。

## 特性

- 支持多种数据库的 SQL 解析和构建：
  - MySQL
  - PostgreSQL
  - SQLite

- SQL 迁移生成：
  - 从 Golang 模型和当前 SQL 架构创建迁移
  - 生成与 `golang-migrate/migrate` 兼容的迁移版本

- 高级功能：
  - 支持嵌入式结构体
  - Avro 架构生成（仅限 MySQL）
  - 与 `gorm` 标签兼容（默认标签为 `sql`）

## 约定

### 默认行为

- 数据库：默认为 `mysql`（使用 `sql_builder.WithPostgresql()` 可切换到 PostgreSQL 等）
- SQL 语法：默认大写（例如：`"SELECT * FROM user WHERE id = ?"`）
  - 使用 `sql_builder.WithSqlLowercase()` 可切换为小写
- 表名：默认单数
  - 使用 `sql_builder.WithPluralTableName()` 可自动添加 's' 实现复数命名
- 注释生成：使用 `sql_builder.WithCommentGenerate()` 选项

### SQL 标签选项

- 格式：支持 `snake_case` 和 `camelCase`（例如：`sql:"primary_key"` 等同于 `sql:"primaryKey"`）
- 自定义列名：`sql:"column:column_name"`
- 主键：`sql:"primary_key"`
- 外键：`sql:"foreign_key:user_id;references:user_id"`
- 自增：`sql:"auto_increment"`
- 默认值：`sql:"default:CURRENT_TIMESTAMP"`
- 覆盖数据类型：`sql:"type:VARCHAR(64)"`
- 忽略字段：`sql:"-"`

### 索引

- 基本索引：`sql:"index"`
- 自定义索引名：`sql:"index:idx_col_name"`
- 唯一索引：`sql:"unique"`
- 自定义唯一索引名：`sql:"unique:idx_name"`
- 复合索引：`sql:"index_columns:col1,col2"`（包括唯一索引和主键）
- 索引类型：`sql:"index_type:btree"`

### 嵌入式结构体

- 使用 `sql:"embedded"` 或 `sql:"squash"`
- 不能是指针
- 支持前缀：`sql:"embedded_prefix:base_"`
- 字段具有最低顺序，除了主键（始终在首位）

### 数据类型

- MySQL 数据类型会被隐式更改：

```sql
TINYINT => tinyint(4)
INT     => int(11)
BIGINT  => bigint(20)
```

### 重要说明

- 指针值必须在结构体中声明

### 示例

1. 使用指针值：

```golang
type sample struct {
    ID        int32 `sql:"primary_key"`
    DeletedAt *time.Time
}

now := time.Now()
newMigration.FromObjects(sample{DeletedAt: &now})
```

2. 嵌入式结构体：

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

3. 完整示例：

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
    //    `id`            int(11) AUTO_INCREMENT PRIMARY KEY,
    //    `alias`         varchar(64),
    //    `name`          varchar(64),
    //    `age`           int(11),
    //    `bio`           text,
    //    `accept_tnc_at` datetime NULL,
    //    `created_at`    datetime DEFAULT CURRENT_TIMESTAMP(),
    //    `updated_at`    datetime DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP()
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