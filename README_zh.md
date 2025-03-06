### SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

[English](README.md) | 中文

SQLize 是一款强大的迁移生成工具，可检测两个 SQL 状态源之间的差异。它通过对比现有 SQL 模式与 Go 模型，简化迁移创建过程，确保数据库平滑更新。
SQLize 设计灵活，支持 `MySQL`、`PostgreSQL` 和 `SQLite`，并能与流行的 Go ORM 和迁移工具（如 `gorm`（gorm 标签）、`golang-migrate/migrate`（迁移版本）等）良好集成。
此外，SQLize 还提供高级功能，包括 `Avro Schema` 导出（仅支持 MySQL）和 `ERD` 关系图生成（`MermaidJS`）。

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

### 使用方法

- 将以下代码添加到您的项目中作为命令。
- 实现 YourModels()，返回受迁移影响的 Go 模型。
- 需要生成迁移时，运行该命令。

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
	sqlLatest := sqlize.NewSqlize(sqlize.WithSqlTag("sql"),
		sqlize.WithMigrationFolder(migrationFolder),
		sqlize.WithCommentGenerate())

	ms := YourModels() // TODO: implement YourModels() function
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

	fmt.Println("\n\n### migration up")
	migrationUp := sqlLatest.StringUp()
	fmt.Println(migrationUp)

	fmt.Println("\n\n### migration down")
	fmt.Println(sqlLatest.StringDown())

	initVersion := false
	if initVersion {
		log.Println("write to init version")
		err = sqlLatest.WriteFilesVersion("new version", 0, false)
		if err != nil {
			log.Fatal("sqlize WriteFilesVersion", err)
		}
	}

	if len(os.Args) > 1 {
		log.Println("write to file", os.Args[1])
		err = sqlLatest.WriteFilesWithVersion(os.Args[1], sqlVersion, false)
		if err != nil {
			log.Fatal("sqlize WriteFilesWithVersion", err)
		}
	}
}
```