package element

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/sunary/sqlize/utils"
)

// Table ...
type Table struct {
	Node

	Columns        []Column
	columnIndexes  map[string]int
	columnPosition *ast.ColumnPosition

	Indexes      []Index
	indexIndexes map[string]int
}

// NewTable ...
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

// AddColumn ...
func (t *Table) AddColumn(col Column) {
	id := t.getIndexColumn(col.Name)
	switch {
	case id == -1:
		t.Columns = append(t.Columns, col)
		t.columnIndexes[col.Name] = len(t.Columns) - 1
		id = len(t.Columns) - 1

	case t.Columns[id].Action != MigrateAddAction:
		t.Columns[id] = col

	default:
		t.Columns[id].Options = append(t.Columns[id].Options, col.Options...)

		if size := len(t.Columns[id].Options); size > 0 {
			for i := range t.Columns[id].Options[:size-1] {
				if t.Columns[id].Options[i].Tp == ast.ColumnOptionPrimaryKey {
					t.Columns[id].Options[i], t.Columns[id].Options[size-1] = t.Columns[id].Options[size-1], t.Columns[id].Options[i]
					break
				}
			}
		}

		t.Columns[id].Typ = col.Typ
		return
	}

	if t.columnPosition != nil {
		defer func() {
			t.columnPosition = nil
		}()

		switch t.columnPosition.Tp {
		case ast.ColumnPositionFirst:
			t.swapOrder(col.Name, id, 0)

		case ast.ColumnPositionAfter:
			if afterID := t.getIndexColumn(t.columnPosition.RelativeColumn.Name.O); afterID >= 0 {
				t.swapOrder(col.Name, id, afterID+1)
			}
		}
	}
}

func (t *Table) swapOrder(colName string, oldID, newID int) {
	if oldID == newID {
		return
	}

	if newID == len(t.Columns)-1 {
		t.Columns = append(t.Columns, t.Columns[oldID])
	} else {
		oldCol := t.Columns[oldID]
		t.Columns = append(t.Columns[:newID+1], t.Columns[newID:]...)
		t.Columns[newID] = oldCol
	}

	switch {
	case oldID > newID:
		for k := range t.columnIndexes {
			if t.columnIndexes[k] >= newID && t.columnIndexes[k] < oldID {
				t.columnIndexes[k]++
			}
		}

		if oldID == len(t.Columns)-2 {
			t.Columns = t.Columns[:len(t.Columns)-1]
		} else {
			t.Columns = append(t.Columns[:oldID], t.Columns[oldID+1:]...)
		}

	case oldID < newID:
		for k := range t.columnIndexes {
			if t.columnIndexes[k] > oldID && t.columnIndexes[k] <= newID {
				t.columnIndexes[k]--
			}
		}

		t.Columns = append(t.Columns[:oldID-1], t.Columns[oldID:]...)
	}

	t.columnIndexes[colName] = newID
}

func (t *Table) removeColumn(colName string) {
	id := t.getIndexColumn(colName)
	switch {
	case id == -1:
		col := Column{Node: Node{Name: colName, Action: MigrateRemoveAction}}
		t.Columns = append(t.Columns, col)
		t.columnIndexes[colName] = len(t.Columns) - 1

	case t.Columns[id].Action == MigrateAddAction:
		t.Columns[id].Action = MigrateNoAction

	default:
		t.Columns[id].Action = MigrateRemoveAction
	}
}

// RenameColumn ...
func (t *Table) RenameColumn(oldName, newName string) {
	if id := t.getIndexColumn(oldName); id >= 0 {
		t.Columns[id].Action = MigrateRenameAction
		t.Columns[id].OldName = oldName
		t.Columns[id].Name = newName
		t.columnIndexes[newName] = id
		delete(t.columnIndexes, oldName)
	}
}

// AddIndex ...
func (t *Table) AddIndex(idx Index) {
	id := t.getIndexColumn(idx.Name)
	if id == -1 {
		t.Indexes = append(t.Indexes, idx)
		t.indexIndexes[idx.Name] = len(t.Indexes) - 1
		return
	}

	t.Indexes[id] = idx
}

// RemoveIndex ...
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

// RenameIndex ...
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
	if v, ok := t.columnIndexes[colName]; ok {
		return v
	}

	return -1
}

func (t Table) getIndexIndex(idxName string) int {
	if v, ok := t.indexIndexes[idxName]; ok {
		return v
	}

	return -1
}

// Diff differ between 2 migrations
func (t *Table) Diff(old Table) {
	for i := range t.Columns {
		if j := old.getIndexColumn(t.Columns[i].Name); t.Columns[i].Action == MigrateAddAction && j >= 0 {
			if (t.Columns[i].Typ != nil && t.Columns[i].Typ.String() == old.Columns[j].Typ.String()) ||
				(t.Columns[i].PgTyp != nil && t.Columns[i].PgTyp == old.Columns[j].PgTyp) {
				t.Columns[i].Action = MigrateNoAction
			} else {
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

type mOrder struct {
	k string
	v int
}

// Arrange correct column order
func (t *Table) Arrange() {
	orders := make([]mOrder, 0, len(t.columnIndexes))
	for k, v := range t.columnIndexes {
		orders = append(orders, mOrder{
			k: k,
			v: v,
		})
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].v < orders[j].v
	})

	for i := range orders {
		for j := range t.Columns {
			if orders[i].k == t.Columns[j].Name {
				t.Columns[i], t.Columns[j] = t.Columns[j], t.Columns[i]
				break
			}
		}
	}
}

// MigrationColumnUp ...
func (t Table) MigrationColumnUp() ([]string, map[string]struct{}) {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		dropCols := make(map[string]struct{})
		for i := range t.Columns {
			if t.Columns[i].Action != MigrateNoAction {
				after := ""
				if t.Columns[i].Action == MigrateAddAction {
					for j := i - 1; j >= 0; j-- {
						if t.Columns[j].Action != MigrateRemoveAction {
							after = t.Columns[j].Name
							break
						}
					}
				} else if t.Columns[i].Action == MigrateRemoveAction {
					dropCols[t.Columns[i].Name] = struct{}{}
				}

				if after != "" {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, after, -1)...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, "", -1)...)
				}
			}
		}

		return strSqls, dropCols

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

		return []string{fmt.Sprintf(sql.CreateTableStm(), utils.EscapeSqlName(sql.IsPostgres, t.Name), strings.Join(strCols, ",\n"), "")}, nil

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(sql.DropTableStm(), utils.EscapeSqlName(sql.IsPostgres, t.Name))}, nil

	case MigrateModifyAction:
		// TODO
		return nil, nil

	default:
		return nil, nil
	}
}

// MigrationIndexUp ...
func (t Table) MigrationIndexUp(dropCols map[string]struct{}) []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Indexes {
			if _, ok := dropCols[t.Indexes[i].Columns[0]]; t.Indexes[i].Action != MigrateNoAction &&
				(t.Indexes[i].Action != MigrateRemoveAction || !ok) {
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

// MigrationColumnDown ...
func (t Table) MigrationColumnDown() ([]string, map[string]struct{}) {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		dropCols := make(map[string]struct{})
		for i := range t.Columns {
			if t.Columns[i].Action != MigrateNoAction {
				after := ""
				if t.Columns[i].Action == MigrateRemoveAction {
					for j := i - 1; j >= 0; j-- {
						if t.Columns[j].Action != MigrateAddAction {
							after = t.Columns[j].Name
							break
						}
					}
				} else if t.Columns[i].Action == MigrateAddAction {
					dropCols[t.Columns[i].Name] = struct{}{}
				}

				if after != "" {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, after)...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, "")...)
				}
			}
		}

		return strSqls, dropCols

	case MigrateAddAction:
		t.Action = MigrateRemoveAction
		return t.MigrationColumnUp()

	case MigrateRemoveAction:
		t.Action = MigrateAddAction
		return t.MigrationColumnUp()

	case MigrateModifyAction:
		// TODO
		return nil, nil

	default:
		return nil, nil
	}
}

// MigrationIndexDown ...
func (t Table) MigrationIndexDown(dropCols map[string]struct{}) []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.Indexes {
			if _, ok := dropCols[t.Indexes[i].Columns[0]]; t.Indexes[i].Action != MigrateNoAction &&
				(t.Indexes[i].Action != MigrateRemoveAction || !ok) {
				strSqls = append(strSqls, t.Indexes[i].migrationDown(t.Name)...)
			}
		}
		return strSqls

	case MigrateAddAction:
		t.Action = MigrateRemoveAction
		return t.MigrationIndexUp(dropCols)

	case MigrateRemoveAction:
		t.Action = MigrateAddAction
		return t.MigrationIndexUp(dropCols)

	case MigrateModifyAction:
		// TODO
		return nil

	default:
		return nil
	}
}
