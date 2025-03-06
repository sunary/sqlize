# SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

English | [中文](README_zh.md)

SQLize is a powerful migration generation tool that detects differences between two SQL state sources. It simplifies migration creation by comparing an existing SQL schema with Go models, ensuring seamless database updates.
Designed for flexibility, SQLize supports `MySQL`, `PostgreSQL`, and `SQLite` and integrates well with popular Go ORM and migration tools like `gorm` (gorm tag), `golang-migrate/migrate` (migration version), and more.
Additionally, SQLize offers advanced features, including `Avro Schema` export (MySQL only) and `ERD` diagram generation (`MermaidJS`).

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

- Pointer values must be declared in the struct or predefined data types.

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

## Usage

- Add the following code to your project as a command.
- Implement `YourModels()` to return the Go models affected by the migration.
- Run the command whenever you need to generate a migration.

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