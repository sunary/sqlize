package element

import (
	"fmt"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/mysql-templates"
	"github.com/sunary/sqlize/utils"
)

type Table struct {
	Node

	Columns        []Column
	columnIndexes  map[string]int
	columnPosition *ast.ColumnPosition

	Indexes      []Index
	indexIndexes map[string]int
}

func NewTable(name string) *Table {
	return &Table{
		Node: Node{
			Name:   name,
			Action: MigrateAddAction,
		},
		Columns:       []Column{},
		columnIndexes: map[string]int{},
		Indexes:       []Index{},
		indexIndexes:  map[string]int{},
	}
}

func (t *Table) AddColumn(col Column) {
	id := t.getIndexColumn(col.Name)
	if id == -1 {
		t.Columns = append(t.Columns, col)
		t.columnIndexes[col.Name] = len(t.Columns) - 1

		if t.columnPosition != nil {
			defer func() {
				t.columnPosition = nil
			}()

			newID, currentID := 0, len(t.Columns)-1
			if t.columnPosition.Tp == ast.ColumnPositionAfter {
				if afterID := t.getIndexColumn(t.columnPosition.RelativeColumn.Name.O); afterID >= 0 {
					newID = afterID + 1
				} else {
					return
				}
			}

			t.Columns[newID], t.Columns[currentID] = t.Columns[currentID], t.Columns[newID]

			for k := range t.columnIndexes {
				if t.columnIndexes[k] >= newID {
					t.columnIndexes[k] += 1
				}
			}
			t.columnIndexes[col.Name] = newID
		}
		return

	}

	if t.Columns[id].Action == MigrateAddAction {
		t.Columns[id].Options = append(t.Columns[id].Options, col.Options...)

		size := len(t.Columns[id].Options)
		for i := range t.Columns[id].Options[:size-1] {
			if t.Columns[id].Options[i].Tp == ast.ColumnOptionPrimaryKey {
				t.Columns[id].Options[i], t.Columns[id].Options[size-1] = t.Columns[id].Options[size-1], t.Columns[id].Options[i]
				break
			}
		}

		t.Columns[id].Typ = col.Typ
	} else {
		t.Columns[id] = col
	}
}

func (t *Table) removeColumn(colName string) {
	id := t.getIndexColumn(colName)
	if id == -1 {
		col := Column{Node: Node{Name: colName, Action: MigrateRemoveAction}}
		t.Columns = append(t.Columns, col)
		t.columnIndexes[colName] = len(t.Columns) - 1
		return
	}

	if t.Columns[id].Action == MigrateAddAction {
		t.Columns[id].Action = MigrateNoAction
	} else {
		t.Columns[id].Action = MigrateRemoveAction
	}
}

func (t *Table) RenameColumn(oldName, newName string) {
	if id := t.getIndexColumn(oldName); id >= 0 {
		t.Columns[id].Action = MigrateRenameAction
		t.Columns[id].OldName = oldName
		t.Columns[id].Name = newName
		t.columnIndexes[newName] = id
		delete(t.columnIndexes, oldName)
	}
}

func (t *Table) AddIndex(idx Index) {
	id := t.getIndexColumn(idx.Name)
	if id == -1 {
		t.Indexes = append(t.Indexes, idx)
		t.indexIndexes[idx.Name] = len(t.Indexes) - 1
		return
	}

	t.Indexes[id] = idx
}

func (t *Table) RemoveIndex(idxName string) {
	id := t.getIndexIndex(idxName)
	if id == -1 {
		idx := Index{Node: Node{Name: idxName, Action: MigrateRemoveAction}}
		t.Indexes = append(t.Indexes, idx)
		t.indexIndexes[idxName] = len(t.Indexes) - 1
		return
	}

	if t.Indexes[id].Action == MigrateAddAction {
		t.Indexes[id].Action = MigrateNoAction
	} else {
		t.Indexes[id].Action = MigrateRemoveAction
	}
}

func (t *Table) RenameIndex(oldName, newName string) {
	if id := t.getIndexIndex(oldName); id >= 0 {
		t.Indexes[id].Action = MigrateRenameAction
		t.Indexes[id].OldName = oldName
		t.Indexes[id].Name = newName
		t.indexIndexes[newName] = id
		delete(t.indexIndexes, oldName)
	}
}

func (t Table) getIndexColumn(colName string) int {
	//if v, ok := t.columnIndexes[colName]; ok {
	//	return v
	//}
	for i := range t.Columns {
		if t.Columns[i].Name == colName {
			return i
		}
	}

	return -1
}

func (t Table) getIndexIndex(idxName string) int {
	//if v, ok := t.indexIndexes[idxName]; ok {
	//	return v
	//}
	for i := range t.Indexes {
		if t.Indexes[i].Name == idxName {
			return i
		}
	}

	return -1
}

func (t *Table) Diff(old Table) {
	for i := range t.Columns {
		if j := old.getIndexColumn(t.Columns[i].Name); t.Columns[i].Action == MigrateAddAction && j >= 0 {
			if t.Columns[i].Typ == old.Columns[j].Typ {
				t.Columns[i].Action = MigrateNoAction
			} else {
				t.Columns[i] = old.Columns[j]
				t.Columns[i].Action = MigrateModifyAction
			}
		}
	}

	for j := range old.Columns {
		if old.Columns[j].Action == MigrateAddAction && t.getIndexColumn(old.Columns[j].Name) == -1 {
			old.Columns[j].Action = MigrateRemoveAction
			t.AddColumn(old.Columns[j])

			newID, currentID := 0, len(t.Columns)-1
			for _, v := range t.columnIndexes {
				if v+1 == j {
					newID = v + 1
					break
				}
			}

			t.Columns[newID], t.Columns[currentID] = t.Columns[currentID], t.Columns[newID]

			for k := range t.columnIndexes {
				if t.columnIndexes[k] >= newID {
					t.columnIndexes[k] += 1
				}
			}
			t.columnIndexes[old.Columns[j].Name] = newID
		}
	}

	for i := range t.Indexes {
		if j := old.getIndexIndex(t.Indexes[i].Name); t.Indexes[i].Action == MigrateAddAction && j >= 0 {
			if t.Indexes[i].Typ == old.Indexes[j].Typ && utils.SlideStrEqual(t.Indexes[i].Columns, old.Indexes[j].Columns) {
				t.Indexes[i].Action = MigrateNoAction
			} else {
				t.Indexes[i] = old.Indexes[j]
				t.Indexes[i].Action = MigrateModifyAction
			}
		}
	}

	for j := range old.Indexes {
		if old.Indexes[j].Action == MigrateAddAction && t.getIndexIndex(old.Indexes[j].Name) == -1 {
			old.Indexes[j].Action = MigrateRemoveAction
			t.AddIndex(old.Indexes[j])
		}
	}
}

func (t Table) MigrationColumnUp() []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Columns {
			if t.Columns[i].Action != MigrateNoAction {
				after := ""
				for k, v := range t.columnIndexes {
					if v == t.columnIndexes[t.Columns[i].Name]-1 {
						after = k
						break
					}
				}

				if after != "" {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, after, -1)...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, "", -1)...)
				}
			}
		}
		return strSqls

	case MigrateAddAction:
		maxIdent := len(t.Columns[0].Name)
		for i := range t.Columns {
			if t.Columns[i].Action == MigrateAddAction || t.Columns[i].Action == MigrateModifyAction || t.Columns[i].Action == MigrateRenameAction {
				if len(t.Columns[i].Name) > maxIdent {
					maxIdent = len(t.Columns[i].Name)
				}
			}
		}

		strCols := make([]string, 0)
		for i := range t.Columns {
			if t.Columns[i].Action == MigrateAddAction {
				strCols = append(strCols, " "+t.Columns[i].migrationUp("", "", maxIdent)[0])
			} else if t.Columns[i].Action == MigrateModifyAction || t.Columns[i].Action == MigrateRenameAction {
				nCol := t.Columns[i]
				nCol.Action = MigrateAddAction
				strCols = append(strCols, " "+nCol.migrationUp("", "", maxIdent)[0])
			}
		}

		return []string{fmt.Sprintf(mysql_templates.CreateTableStm(isLower), utils.EscapeSqlName(t.Name), strings.Join(strCols, ",\n"))}

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(mysql_templates.DropTableStm(isLower), utils.EscapeSqlName(t.Name))}

	case MigrateModifyAction:
		// TODO
		return nil

	default:
		return nil
	}
}

func (t Table) MigrationIndexUp() []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Indexes {
			if t.Indexes[i].Action != MigrateNoAction {
				strSqls = append(strSqls, t.Indexes[i].migrationUp(t.Name)...)
			}
		}
		return strSqls

	case MigrateAddAction:
		strSqls := make([]string, 0)
		for i := range t.Indexes {
			if t.Indexes[i].Action == MigrateAddAction {
				strSqls = append(strSqls, t.Indexes[i].migrationUp(t.Name)...)
			}
		}
		return strSqls

	case MigrateRemoveAction:
		return nil

	case MigrateModifyAction:
		return nil

	default:
		return nil
	}
}

func (t Table) MigrationColumnDown() []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Columns {
			if t.Columns[i].Action != MigrateNoAction {
				after := ""
				for k, v := range t.columnIndexes {
					if v == t.columnIndexes[t.Columns[i].Name]-1 {
						after = k
						break
					}
				}

				if after != "" {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, after)...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, "")...)
				}
			}
		}
		return strSqls

	case MigrateAddAction:
		t.Action = MigrateRemoveAction
		return t.MigrationColumnUp()

	case MigrateRemoveAction:
		t.Action = MigrateAddAction
		return t.MigrationColumnUp()

	case MigrateModifyAction:
		// TODO
		return nil

	default:
		return nil
	}
}

func (t Table) MigrationIndexDown() []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Indexes {
			if t.Indexes[i].Action != MigrateNoAction {
				strSqls = append(strSqls, t.Indexes[i].migrationDown(t.Name)...)
			}
		}
		return strSqls

	case MigrateAddAction:
		t.Action = MigrateRemoveAction
		return t.MigrationIndexUp()

	case MigrateRemoveAction:
		t.Action = MigrateAddAction
		return t.MigrationIndexUp()

	case MigrateModifyAction:
		// TODO
		return nil

	default:
		return nil
	}
}
