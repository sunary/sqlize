package element

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	ptypes "github.com/auxten/postgresql-parser/pkg/sql/types"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/types"
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

	ForeignKeys      []ForeignKey
	indexForeignKeys map[string]int
}

// NewTable ...
func NewTable(name string) *Table {
	return NewTableWithAction(name, MigrateAddAction)
}

// NewTableWithAction ...
func NewTableWithAction(name string, action MigrateAction) *Table {
	return &Table{
		Node: Node{
			Name:   name,
			Action: action,
		},
		Columns:          []Column{},
		columnIndexes:    map[string]int{},
		Indexes:          []Index{},
		indexIndexes:     map[string]int{},
		ForeignKeys:      []ForeignKey{},
		indexForeignKeys: map[string]int{},
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
		t.Columns[id].CurrentAttr.Options = append(t.Columns[id].CurrentAttr.Options, col.CurrentAttr.Options...)

		if size := len(t.Columns[id].CurrentAttr.Options); size > 0 {
			for i := range t.Columns[id].CurrentAttr.Options[:size-1] {
				if t.Columns[id].CurrentAttr.Options[i].Tp == ast.ColumnOptionPrimaryKey {
					t.Columns[id].CurrentAttr.Options[i], t.Columns[id].CurrentAttr.Options[size-1] = t.Columns[id].CurrentAttr.Options[size-1], t.Columns[id].CurrentAttr.Options[i]
					break
				}
			}
		}

		t.Columns[id].CurrentAttr.MysqlType = col.CurrentAttr.MysqlType
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
	id := t.getIndexIndex(idx.Name)
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

// AddForeignKey ...
func (t *Table) AddForeignKey(fk ForeignKey) {
	id := t.getIndexForeignKey(fk.Name)
	if id == -1 {
		t.ForeignKeys = append(t.ForeignKeys, fk)
		t.indexForeignKeys[fk.Name] = len(t.ForeignKeys) - 1
	} else {
		t.ForeignKeys[id] = fk
	}

	for i := range t.Columns {
		if t.Columns[i].Name == fk.Column {
			t.Columns[i].CurrentAttr.Options = append(t.Columns[i].CurrentAttr.Options, &ast.ColumnOption{
				Tp: ast.ColumnOptionReference,
			})
		}
	}
}

// RemoveForeignKey ...
func (t *Table) RemoveForeignKey(fkName string) {
	id := t.getIndexForeignKey(fkName)
	if id == -1 {
		fk := ForeignKey{Node: Node{Name: fkName, Action: MigrateRemoveAction}}
		t.ForeignKeys = append(t.ForeignKeys, fk)
		t.indexForeignKeys[fkName] = len(t.ForeignKeys) - 1
		return
	}

	if t.ForeignKeys[id].Action == MigrateAddAction {
		t.ForeignKeys[id].Action = MigrateNoAction
	} else {
		t.ForeignKeys[id].Action = MigrateRemoveAction
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

func (t Table) getIndexForeignKey(fkName string) int {
	if v, ok := t.indexForeignKeys[fkName]; ok {
		return v
	}

	return -1
}

func hasChangedMysqlOptions(new, old []*ast.ColumnOption) bool {
	optNew := make([]*ast.ColumnOption, 0, len(new))
	hasComment := false
	for _, v := range new {
		if !hasComment {
			optNew = append(optNew, v)
		}
		if v.Tp == ast.ColumnOptionComment {
			hasComment = true
		}
	}

	optOld := make([]*ast.ColumnOption, 0, len(old))
	hasComment = false
	for _, v := range old {
		if !hasComment {
			optOld = append(optOld, v)
		}
		if v.Tp == ast.ColumnOptionComment {
			hasComment = true
		}
	}

	if len(optNew) != len(optOld) {
		return true
	}

	mNew := map[ast.ColumnOptionType]int{}
	for i := range optNew {
		mNew[optNew[i].Tp] += 1
	}

	mOld := map[ast.ColumnOptionType]int{}
	for i := range optOld {
		mOld[optOld[i].Tp] += 1
	}

	for k, v := range mOld {
		if mNew[k] != v {
			return true
		}
	}

	return false
}

func hasChangedMysqlType(new, old *types.FieldType) bool {
	return new != nil && new.String() != old.String()
}

func hasChangePostgresType(new, old *ptypes.T) bool {
	return new != nil && new.SQLString() != old.SQLString()
}

// Diff differ between 2 migrations
func (t *Table) Diff(old Table) {
	for i := range t.Columns {
		if j := old.getIndexColumn(t.Columns[i].Name); t.Columns[i].Action == MigrateAddAction &&
			j >= 0 && old.Columns[j].Action != MigrateNoAction {
			if hasChangedMysqlOptions(t.Columns[i].CurrentAttr.Options, old.Columns[j].CurrentAttr.Options) ||
				hasChangedMysqlType(t.Columns[i].CurrentAttr.MysqlType, old.Columns[j].CurrentAttr.MysqlType) ||
				hasChangePostgresType(t.Columns[i].CurrentAttr.PgType, old.Columns[j].CurrentAttr.PgType) {
				t.Columns[i].Action = MigrateModifyAction
				t.Columns[i].PreviousAttr = old.Columns[j].CurrentAttr
			} else {
				t.Columns[i].Action = MigrateNoAction
			}
		}
	}

	for j := range old.Columns {
		if old.Columns[j].Action == MigrateAddAction && t.getIndexColumn(old.Columns[j].Name) == -1 {
			old.Columns[j].Action = MigrateRemoveAction
			t.AddColumn(old.Columns[j])

			if j == 0 {
				t.swapOrder(old.Columns[j].Name, len(t.Columns)-1, 0)
			} else if newID, ok := old.columnIndexes[old.Columns[j-1].Name]; ok {
				t.swapOrder(old.Columns[j].Name, len(t.Columns)-1, newID+1)
			}
		}
	}

	for i := range t.Indexes {
		if j := old.getIndexIndex(t.Indexes[i].Name); t.Indexes[i].Action == MigrateAddAction &&
			j >= 0 && old.Indexes[j].Action != MigrateNoAction {
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

	for i := range t.ForeignKeys {
		if j := old.getIndexForeignKey(t.ForeignKeys[i].Name); t.ForeignKeys[i].Action == MigrateAddAction &&
			j >= 0 && old.ForeignKeys[j].Action != MigrateNoAction {
			t.ForeignKeys[i] = old.ForeignKeys[j]
			t.ForeignKeys[i].Action = MigrateModifyAction
		}
	}

	for j := range old.ForeignKeys {
		if old.ForeignKeys[j].Action == MigrateAddAction && t.getIndexForeignKey(old.ForeignKeys[j].Name) == -1 {
			old.ForeignKeys[j].Action = MigrateRemoveAction
			t.AddForeignKey(old.ForeignKeys[j])
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

func (t Table) hashValue() string {
	cols, idxs := make([]string, len(t.Columns)), make([]string, len(t.Indexes))
	for i := range t.Columns {
		cols[i] = t.Columns[i].hashValue()
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})

	for i := range t.Indexes {
		idxs[i] = t.Indexes[i].hashValue()
	}
	sort.Slice(idxs, func(i, j int) bool {
		return idxs[i] < idxs[j]
	})

	strHash := strings.Join(append(cols, idxs...), ";")
	hash := md5.Sum([]byte(strHash))
	return hex.EncodeToString(hash[:])
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

				if ignoreFieldOrder || after == "" {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, "", -1)...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationUp(t.Name, after, -1)...)
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
		commentCols := make([]string, 0)
		for i := range t.Columns {
			switch t.Columns[i].Action {
			case MigrateAddAction:
				strCols = append(strCols, " "+t.Columns[i].migrationUp("", "", maxIdent)[0])
			case MigrateModifyAction, MigrateRenameAction:
				nCol := t.Columns[i]
				nCol.Action = MigrateAddAction
				strCols = append(strCols, " "+nCol.migrationUp("", "", maxIdent)[0])
			}

			commentCols = append(commentCols, t.Columns[i].migrationCommentUp(t.Name)...)
		}

		return append([]string{fmt.Sprintf(sql.CreateTableStm(), sql.EscapeSqlName(t.Name), strings.Join(strCols, ",\n"), "")}, commentCols...), nil

	case MigrateRemoveAction:
		return []string{fmt.Sprintf(sql.DropTableStm(), sql.EscapeSqlName(t.Name))}, nil

	case MigrateModifyAction:
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

// MigrationForeignKeyUp ...
func (t Table) MigrationForeignKeyUp(dropCols map[string]struct{}) []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.ForeignKeys {
			if _, ok := dropCols[t.ForeignKeys[i].Column]; t.ForeignKeys[i].Action != MigrateNoAction &&
				(t.ForeignKeys[i].Action != MigrateRemoveAction || !ok) {
				strSqls = append(strSqls, t.ForeignKeys[i].migrationUp(t.Name)...)
			}
		}
		return strSqls

	case MigrateAddAction:
		strSqls := make([]string, 0)
		for i := range t.ForeignKeys {
			if t.ForeignKeys[i].Action == MigrateAddAction {
				strSqls = append(strSqls, t.ForeignKeys[i].migrationUp(t.Name)...)
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

				if ignoreFieldOrder || after == "" {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, "")...)
				} else {
					strSqls = append(strSqls, t.Columns[i].migrationDown(t.Name, after)...)
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
		return nil

	default:
		return nil
	}
}

// MigrationForeignKeyDown ...
func (t Table) MigrationForeignKeyDown(dropCols map[string]struct{}) []string {
	switch t.Action {
	case MigrateNoAction:
		strSqls := make([]string, 0)
		for i := range t.ForeignKeys {
			if _, ok := dropCols[t.ForeignKeys[i].Column]; t.ForeignKeys[i].Action != MigrateNoAction &&
				(t.ForeignKeys[i].Action != MigrateRemoveAction || !ok) {
				strSqls = append(strSqls, t.ForeignKeys[i].migrationDown(t.Name)...)
			}
		}
		return strSqls

	case MigrateAddAction:
		t.Action = MigrateRemoveAction
		return t.MigrationForeignKeyUp(dropCols)

	case MigrateRemoveAction:
		t.Action = MigrateAddAction
		return t.MigrationForeignKeyUp(dropCols)

	case MigrateModifyAction:
		return nil

	default:
		return nil
	}
}
