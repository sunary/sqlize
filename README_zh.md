# SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

[English](README.md) | 中文

## 项目目的

**SQLize** 通过比较两个 schema 来源生成数据库迁移：你的 **Go 结构体**（期望状态）和 **已有迁移文件**（当前状态）。你只需在 Go 中定义模型，SQLize 会自动生成所需的 `ALTER TABLE`、`CREATE TABLE` 等语句，使数据库与模型保持同步。

### 解决什么问题？

- **手写迁移容易出错** — 手写 `ALTER TABLE` 时容易遗漏列、索引或外键
- **Schema 漂移** — Go 模型与数据库会随时间不同步
- **重复劳动** — 每次 schema 变更都要手动编写 up/down 迁移

### 工作原理

1. **期望状态** — 从 Go 结构体加载 schema（通过 `FromObjects`）
2. **当前状态** — 从迁移目录加载 schema（通过 `FromMigrationFolder`）
3. **差异计算** — SQLize 比较两者并计算变更
4. **输出** — 获取迁移 SQL（`StringUp` / `StringDown`）或直接写入文件（`WriteFilesWithVersion`）

```
Go 结构体（期望）  ──┐
                    ├──► Diff ──► 迁移 SQL（up/down）
迁移文件（当前）  ──┘
```

## 功能特性

- **多数据库支持**：MySQL、PostgreSQL、SQLite、SQL Server
- **ORM 友好**：支持结构体标签（`sql`、`gorm`），兼容 `golang-migrate/migrate`
- **Schema 导出**：Avro Schema（MySQL）、ERD 图（MermaidJS）

## 安装

```bash
go get github.com/sunary/sqlize
```

## 快速开始

```golang
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sunary/sqlize"
)

func main() {
	migrationFolder := "migrations/"
	sqlLatest := sqlize.NewSqlize(
		sqlize.WithSqlTag("sql"),
		sqlize.WithMigrationFolder(migrationFolder),
		sqlize.WithCommentGenerate(),
	)

	ms := YourModels() // 返回你的 Go 模型
	err := sqlLatest.FromObjects(ms...)
	if err != nil {
		log.Fatal("sqlize FromObjects", err)
	}
	sqlVersion := sqlLatest.HashValue()

	sqlMigrated := sqlize.NewSqlize(sqlize.WithMigrationFolder(migrationFolder))
	err = sqlMigrated.FromMigrationFolder()
	if err != nil {
		log.Fatal("sqlize FromMigrationFolder", err)
	}

	sqlLatest.Diff(*sqlMigrated)

	fmt.Println("sql version", sqlVersion)
	fmt.Println("\n### migration up")
	fmt.Println(sqlLatest.StringUp())
	fmt.Println("\n### migration down")
	fmt.Println(sqlLatest.StringDown())

	if len(os.Args) > 1 {
		err = sqlLatest.WriteFilesWithVersion(os.Args[1], sqlVersion, false)
		if err != nil {
			log.Fatal("sqlize WriteFilesWithVersion", err)
		}
	}
}
```

## 约定

### 默认行为

- 数据库：`mysql`（使用 `sqlize.WithPostgresql()`、`sqlize.WithSqlite()` 等切换）
- SQL 语法：大写（使用 `sqlize.WithSqlLowercase()` 切换为小写）
- 表名：单数（使用 `sqlize.WithPluralTableName()` 切换为复数）
- 注释生成：`sqlize.WithCommentGenerate()`

### SQL 标签选项

- 格式：支持 `snake_case` 和 `camelCase`（例如 `sql:"primary_key"` 等同于 `sql:"primaryKey"`）
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
- 字段具有最低顺序，主键除外（始终在首位）

### 数据类型

- MySQL 数据类型会被隐式更改：

```sql
TINYINT => tinyint(4)
INT     => int(11)
BIGINT  => bigint(20)
```

- 指针值必须在结构体或预定义数据类型中声明：

```golang
// your struct
type Record struct {
	ID        int
	DeletedAt *time.Time
}

// =>
// the struct is declared with a value
now := time.Now()
Record{DeletedAt: &now}

// or predefined data type
type Record struct {
	ID        int
	DeletedAt *time.Time `sql:"type:DATETIME"`
}

// or using struct supported by "database/sql"
type Record struct {
	ID        int
	DeletedAt sql.NullTime
}
```
