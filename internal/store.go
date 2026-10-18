package internal

// Store reads and writes PATH values. The registry implementation is only
// used by main; tests inject a fake.
type Store interface {
	ReadPATH(scope Scope) ([]string, error)
	ReadRaw(scope Scope) (string, error)
	WritePATH(scope Scope, entries []string) error
	Broadcast()
}

// RegistryStore talks to the live Windows registry.
type RegistryStore struct{}

// NewStore returns the live registry-backed store.
func NewStore() Store {
	return RegistryStore{}
}

func (RegistryStore) ReadPATH(scope Scope) ([]string, error) {
	return ReadPATH(scope)
}

func (RegistryStore) ReadRaw(scope Scope) (string, error) {
	return ReadRaw(scope)
}

func (RegistryStore) WritePATH(scope Scope, entries []string) error {
	return WritePATH(scope, entries)
}

func (RegistryStore) Broadcast() {
	BroadcastChange()
}
