package element

type MigrateAction int8

const (
	MigrateNoAction MigrateAction = iota
	MigrateAddAction
	MigrateRemoveAction
	MigrateModifyAction
	MigrateRenameAction
)

type Node struct {
	Name    string
	OldName string
	Action  MigrateAction
}
