package element

type MigrateAction int8

const (
	MigrateNoAction MigrateAction = iota
	MigrateAddAction
	MigrateRemoveAction
	MigrateModifyAction
	MigrateRenameAction
)

// Node primitive element
type Node struct {
	Name    string
	OldName string
	Action  MigrateAction
}
