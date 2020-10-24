package element

import (
	"testing"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/model"
)

func TestGetIndexIndex(t *testing.T) {
	const (
		newIndex  = "newIndex"
		newIndex2 = "newIndex2"
	)
	table := NewTable("testTable")

	index := Index{
		Node: Node{
			Name: newIndex,
		},
	}

	index2 := Index{
		Node: Node{
			Name: newIndex2,
		},
	}

	table.AddIndex(index)
	table.AddIndex(index2)

	if i := table.getIndexIndex(newIndex); i != 0 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexIndex(newIndex2); i != 1 {
		t.Errorf("wrong index returned: %d", i)
	}
}

func TestGetIndexColumn(t *testing.T) {
	const (
		newColumn  = "newColumn"
		newColumn2 = "newColumn2"
	)
	table := NewTable("testTable")

	column := Column{
		Node: Node{
			Name: newColumn,
		},
	}

	column2 := Column{
		Node: Node{
			Name: newColumn2,
		},
	}

	table.AddColumn(column)
	table.AddColumn(column2)

	if i := table.getIndexColumn(newColumn); i != 0 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn2); i != 1 {
		t.Errorf("wrong index returned: %d", i)
	}
}

func TestGetIndexColumnWithColumnPositionNone(t *testing.T) {
	const (
		newColumn  = "newColumn"
		newColumn2 = "newColumn2"
		newColumn3 = "newColumn3"
	)

	table := NewTable("testTable")

	column := Column{
		Node: Node{
			Name: newColumn,
		},
	}

	column2 := Column{
		Node: Node{
			Name: newColumn2,
		},
	}

	column3 := Column{
		Node: Node{
			Name: newColumn3,
		},
	}

	table.AddColumn(column)
	table.AddColumn(column2)

	table.columnPosition = &ast.ColumnPosition{
		Tp: ast.ColumnPositionNone,
	}
	table.AddColumn(column3)

	if i := table.getIndexColumn(newColumn); i != 0 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn2); i != 1 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn3); i != 2 {
		t.Errorf("wrong index returned: %d", i)
	}
}

func TestGetIndexColumnWithColumnPositionAfter(t *testing.T) {
	const (
		newColumn  = "newColumn"
		newColumn2 = "newColumn2"
		newColumn3 = "newColumn3"
	)

	table := NewTable("testTable")

	column := Column{
		Node: Node{
			Name: newColumn,
		},
	}

	column2 := Column{
		Node: Node{
			Name: newColumn2,
		},
	}

	column3 := Column{
		Node: Node{
			Name: newColumn3,
		},
	}

	table.AddColumn(column)
	table.AddColumn(column2)

	table.columnPosition = &ast.ColumnPosition{
		Tp: ast.ColumnPositionAfter,
		RelativeColumn: &ast.ColumnName{
			Name: model.CIStr{
				O: newColumn,
			},
		},
	}
	table.AddColumn(column3)

	if i := table.getIndexColumn(newColumn); i != 0 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn2); i != 2 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn3); i != 1 {
		t.Errorf("wrong index returned: %d", i)
	}
}

func TestGetIndexColumnWithColumnPositionFirst(t *testing.T) {
	const (
		newColumn  = "newColumn"
		newColumn2 = "newColumn2"
		newColumn3 = "newColumn3"
	)

	table := NewTable("testTable")

	column := Column{
		Node: Node{
			Name: newColumn,
		},
	}

	column2 := Column{
		Node: Node{
			Name: newColumn2,
		},
	}

	column3 := Column{
		Node: Node{
			Name: newColumn3,
		},
	}

	table.AddColumn(column)
	table.AddColumn(column2)

	table.columnPosition = &ast.ColumnPosition{
		Tp: ast.ColumnPositionFirst,
	}
	table.AddColumn(column3)

	if i := table.getIndexColumn(newColumn); i != 2 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn2); i != 1 {
		t.Errorf("wrong index returned: %d", i)
	}

	if i := table.getIndexColumn(newColumn3); i != 0 {
		t.Errorf("wrong index returned: %d", i)
	}
}
