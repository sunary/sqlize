package element

import (
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/sql-templates"
)

var (
	sql *sql_templates.Sql
)

type Migration struct {
	currentTable string
	Tables       []Table
	tableIndexes map[string]int
}

func NewMigration(isPostgres, isLower bool) Migration {
	sql = sql_templates.NewSql(isPostgres, isLower)
	return Migration{
		Tables:       []Table{},
		tableIndexes: map[string]int{},
	}
}

func (m *Migration) Using(tbName string) {
	if tbName != "" {
		m.currentTable = tbName
	}
}

func (m *Migration) AddTable(tb Table) {
	id := m.getIndexTable(tb.Name)
	if id == -1 {
		m.Tables = append(m.Tables, tb)
		m.tableIndexes[tb.Name] = len(m.Tables) - 1
		return
	}

	m.Tables[id] = tb
}

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

func (m *Migration) RenameTable(oldName, newName string) {
	if id := m.getIndexTable(oldName); id >= 0 {
		m.Tables[id].Action = MigrateRemoveAction
		m.Tables[id].OldName = oldName
		m.Tables[id].Name = newName
		m.tableIndexes[newName] = id
		delete(m.tableIndexes, oldName)
	}
}

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

func (m *Migration) SetColumnPosition(tbName string, pos *ast.ColumnPosition) {
	if tbName == "" {
		tbName = m.currentTable
	}

	if id := m.getIndexTable(tbName); id >= 0 {
		m.Tables[id].columnPosition = pos
	}
}

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

func (m *Migration) RenameColumn(tbName, oldName, newName string) {
	if tbName == "" {
		tbName = m.currentTable
	}

	if id := m.getIndexTable(tbName); id >= 0 {
		m.Tables[id].RenameColumn(oldName, newName)
	}
}

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

func (m Migration) MigrationUp() string {
	strTables := make([]string, 0)
	for i := range m.Tables {
		strTb := make([]string, 0)
		if mColumn := m.Tables[i].MigrationColumnUp(); len(mColumn) > 0 {
			strTb = append(strTb, strings.Join(mColumn, "\n"))
		}
		if mIndex := m.Tables[i].MigrationIndexUp(); len(mIndex) > 0 {
			strTb = append(strTb, strings.Join(mIndex, "\n"))
		}

		if len(strTb) > 0 {
			strTables = append(strTables, strings.Join(strTb, "\n"))
		}
	}

	return strings.Join(strTables, "\n\n")
}

func (m Migration) MigrationDown() string {
	strTables := make([]string, 0)
	for i := range m.Tables {
		strTb := make([]string, 0)
		if mIndex := m.Tables[i].MigrationIndexDown(); len(mIndex) > 0 {
			strTb = append(strTb, strings.Join(mIndex, "\n"))
		}
		if mColumn := m.Tables[i].MigrationColumnDown(); len(mColumn) > 0 {
			strTb = append(strTb, strings.Join(mColumn, "\n"))
		}

		if len(strTb) > 0 {
			strTables = append(strTables, strings.Join(strTb, "\n"))
		}
	}

	return strings.Join(strTables, "\n\n")
}
