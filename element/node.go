package element

// MigrateAction ...
type MigrateAction int8

const (
	// MigrateNoAction ...
	MigrateNoAction MigrateAction = iota
	// MigrateAddAction ...
	MigrateAddAction
	// MigrateRemoveAction ...
	MigrateRemoveAction
	// MigrateModifyAction ...
	MigrateModifyAction
	// MigrateRenameAction ...
	MigrateRenameAction
)

// Node primitive element
type Node struct {
	Name    string
	OldName string
	Action  MigrateAction
}
