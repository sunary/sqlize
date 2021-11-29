package element

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/sql-templates"
)

var (
	sql *sql_templates.Sql
)

// Migration ...
type Migration struct {
	currentTable string
	Tables       []Table
	tableIndexes map[string]int
}

// NewMigration ...
func NewMigration(isPostgres, isLower bool) Migration {
	sql = sql_templates.NewSql(isPostgres, isLower)
	return Migration{
		Tables:       []Table{},
		tableIndexes: map[string]int{},
	}
}

// Using set current table
func (m *Migration) Using(tbName string) {
	if tbName != "" {
		m.currentTable = tbName
	}
}

// AddTable ...
func (m *Migration) AddTable(tb Table) {
	id := m.getIndexTable(tb.Name)
	if id == -1 {
		m.Tables = append(m.Tables, tb)
		m.tableIndexes[tb.Name] = len(m.Tables) - 1
		return
	}

	m.Tables[id] = tb
}

// RemoveTable ...
func (m *Migration) RemoveTable(tbName string) {
	id := m.getIndexTable(tbName)
	if id == -1 {
		tb := NewTable(tbName)
		tb.Action = MigrateRemoveAction
		m.Tables = append(m.Tables, *tb)
		m.tableIndexes[tb.Name] = len(m.Tables) - 1
		return
	}

	if m.Tables[id].Action == MigrateAddAction {
		m.Tables[id].Action = MigrateNoAction
	} else {
		m.Tables[id].Action = MigrateRemoveAction
	}
}

// RenameTable ...
func (m *Migration) RenameTable(oldName, newName string) {
	if id := m.getIndexTable(oldName); id >= 0 {
		m.Tables[id].Action = MigrateRemoveAction
		m.Tables[id].OldName = oldName
		m.Tables[id].Name = newName
		m.tableIndexes[newName] = id
		delete(m.tableIndexes, oldName)
	}
}

// AddColumn ...
func (m *Migration) AddColumn(tbName string, col Column) {
	if tbName == "" {
		tbName = m.currentTable
	}

	id := m.getIndexTable(tbName)
	if id == -1 {
		tb := NewTable(tbName)
		tb.Action = MigrateNoAction
		m.AddTable(*tb)
		id = len(m.Tables) - 1
	}

	m.Tables[id].AddColumn(col)
}

// SetColumnPosition ...
func (m *Migration) SetColumnPosition(tbName string, pos *ast.ColumnPosition) {
	if tbName == "" {
		tbName = m.currentTable
	}

	if id := m.getIndexTable(tbName); id >= 0 {
		m.Tables[id].columnPosition = pos
	}
}

// RemoveColumn ...
func (m *Migration) RemoveColumn(tbName, colName string) {
	if tbName == "" {
		tbName = m.currentTable
	}

	id := m.getIndexTable(tbName)
	if id == -1 {
		m.AddTable(*NewTable(tbName))
		id = len(m.Tables) - 1
	}

	m.Tables[id].removeColumn(colName)
}

// RenameColumn ...
func (m *Migration) RenameColumn(tbName, oldName, newName string) {
	if tbName == "" {
		tbName = m.currentTable
	}

	if id := m.getIndexTable(tbName); id >= 0 {
		m.Tables[id].RenameColumn(oldName, newName)
	}
}

// AddIndex ...
func (m *Migration) AddIndex(tbName string, idx Index) {
	if tbName == "" {
		tbName = m.currentTable
	}

	id := m.getIndexTable(tbName)
	if id == -1 {
		m.AddTable(*NewTable(tbName))
		id = len(m.Tables) - 1
	}

	m.Tables[id].AddIndex(idx)
}

// RemoveIndex ...
func (m *Migration) RemoveIndex(tbName string, idxName string) {
	if tbName == "" {
		tbName = m.currentTable
	}

	id := m.getIndexTable(tbName)
	if id == -1 {
		m.AddTable(*NewTable(tbName))
		id = len(m.Tables) - 1
	}

	m.Tables[id].RemoveIndex(idxName)
}

// RenameIndex ...
func (m *Migration) RenameIndex(tbName, oldName, newName string) {
	if tbName == "" {
		tbName = m.currentTable
	}

	if id := m.getIndexTable(tbName); id >= 0 {
		m.Tables[id].RenameIndex(oldName, newName)
	}
}

func (m Migration) getIndexTable(tableName string) int {
	if v, ok := m.tableIndexes[tableName]; ok {
		return v
	}

	return -1
}

// Diff differ between 2 migrations
func (m *Migration) Diff(old Migration) {
	for i := range m.Tables {
		if j := old.getIndexTable(m.Tables[i].Name); j >= 0 {
			m.Tables[i].Diff(old.Tables[j])
			m.Tables[i].Action = MigrateNoAction
		}
	}

	for j := range old.Tables {
		if i := m.getIndexTable(old.Tables[j].Name); i == -1 {
			old.Tables[j].Action = MigrateRemoveAction
			m.AddTable(old.Tables[j])
		}
	}
}

// HashValue ...
func (m Migration) HashValue() string {
	tbs := make([]string, len(m.Tables))
	for i := range m.Tables {
		tbs[i] = m.Tables[i].hashValue()
	}

	strHash := strings.Join(tbs, " ")
	hash := md5.Sum([]byte(strHash))
	return hex.EncodeToString(hash[:])
}

// MigrationUp ...
func (m Migration) MigrationUp() string {
	strTables := make([]string, 0)
	for i := range m.Tables {
		m.Tables[i].Arrange()

		strTb := make([]string, 0)
		mCols, dropCols := m.Tables[i].MigrationColumnUp()
		if len(mCols) > 0 {
			strTb = append(strTb, strings.Join(mCols, "\n"))
		}
		if mIdxs := m.Tables[i].MigrationIndexUp(dropCols); len(mIdxs) > 0 {
			strTb = append(strTb, strings.Join(mIdxs, "\n"))
		}

		if len(strTb) > 0 {
			strTables = append(strTables, strings.Join(strTb, "\n"))
		}
	}

	return strings.Join(strTables, "\n\n")
}

// MigrationDown ...
func (m Migration) MigrationDown() string {
	strTables := make([]string, 0)
	for i := range m.Tables {
		m.Tables[i].Arrange()

		strTb := make([]string, 0)
		mCols, dropCols := m.Tables[i].MigrationColumnDown()
		if len(mCols) > 0 {
			strTb = append(strTb, strings.Join(mCols, "\n"))
		}
		if mIdxs := m.Tables[i].MigrationIndexDown(dropCols); len(mIdxs) > 0 {
			strTb = append(strTb, strings.Join(mIdxs, "\n"))
		}

		if len(strTb) > 0 {
			strTables = append(strTables, strings.Join(strTb, "\n"))
		}
	}

	return strings.Join(strTables, "\n\n")
}
