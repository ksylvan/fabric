package db

// Storage is the contract for a named-entity store. Every method that
// takes a name requires one that satisfies fsdb.ValidateStorageName;
// implementations reject other names with *fsdb.InvalidStorageNameError,
// which HTTP handlers map to 400.
type Storage[T any] interface {
	Configure() (err error)
	Get(name string) (ret *T, err error)
	GetNames() (ret []string, err error)
	Delete(name string) (err error)
	// Exists reports false for an invalid name; it cannot distinguish
	// "rejected name" from "absent entry".
	Exists(name string) (ret bool)
	Rename(oldName, newName string) (err error)
	Save(name string, content []byte) (err error)
	Load(name string) (ret []byte, err error)
	ListNames(shellCompleteList bool) (err error)
}
