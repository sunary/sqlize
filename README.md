# SQLize

![github action](https://github.com/sunary/sqlize/actions/workflows/go.yml/badge.svg)

English | [中文](README_zh.md)

## Purpose

**SQLize** generates database migrations by comparing two schema sources: your **Go structs** (desired state) and your **existing migrations** (current state). Instead of writing migration SQL by hand, you define models in Go and SQLize produces the `ALTER TABLE`, `CREATE TABLE`, and related statements needed to bring your database up to date.

### What problem does it solve?

- **Manual migrations are error-prone** — Easy to forget columns, indexes, or foreign keys when writing `ALTER TABLE` by hand
- **Schema drift** — Go models and the database can get out of sync over time
- **Boilerplate** — Repetitive work creating up/down migrations for every schema change

### How it works

1. **Desired state** — Load schema from your Go structs (via `FromObjects`)
2. **Current state** — Load schema from your migration folder (via `FromMigrationFolder`)
3. **Diff** — SQLize compares them and computes the changes
4. **Output** — Get migration SQL (`StringUp` / `StringDown`) or write files directly (`WriteFilesWithVersion`)

```
Go structs (desired)  ──┐
                       ├──► Diff ──► Migration SQL (up/down)
Migration files (current) ─┘
```

## Features

- **Multi-database**: MySQL, PostgreSQL, SQLite, SQL Server
- **ORM-friendly**: Works with struct tags (`sql`, `gorm`), compatible with `golang-migrate/migrate`
- **Schema export**: Avro Schema (MySQL), ERD diagrams (MermaidJS)

## Installation

```bash
go get github.com/sunary/sqlize
```

## Quick Start

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

	ms := YourModels() // Return your Go models
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

## Conventions

### Default Behaviors

- Database: `mysql` (use `sqlize.WithPostgresql()`, `sqlize.WithSqlite()`, etc.)
- SQL syntax: Uppercase (use `sqlize.WithSqlLowercase()` for lowercase)
- Table naming: Singular (use `sqlize.WithPluralTableName()` for plural)
- Comment generation: `sqlize.WithCommentGenerate()`

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

- Pointer values must be declared in the struct or predefined data types:

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
