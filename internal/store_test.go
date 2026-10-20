package internal

import "testing"

func TestRegistryStoreDelegatesToRegistryLayer(t *testing.T) {
	key := &fakeRegistryKey{value: `C:\Windows;%USERPROFILE%\bin`}
	installFakeOpener(t, &fakeOpener{key: key})

	store := NewStore()

	entries, err := store.ReadPATH(ScopeUser)
	if err != nil {
		t.Fatalf("ReadPATH(User) = %v", err)
	}
	if len(entries) != 2 || entries[0] != `C:\Windows` {
		t.Fatalf("ReadPATH(User) = %#v", entries)
	}

	raw, err := store.ReadRaw(ScopeUser)
	if err != nil {
		t.Fatalf("ReadRaw(User) = %v", err)
	}
	if raw != key.value {
		t.Fatalf("ReadRaw(User) = %q, want %q", raw, key.value)
	}

	if err := store.WritePATH(ScopeSystem, []string{`C:\Tools`}); err != nil {
		t.Fatalf("WritePATH(System) = %v", err)
	}
	if key.setValue != `C:\Tools` {
		t.Fatalf("written value = %q, want C:\\\\Tools", key.setValue)
	}

	if _, ok := store.(RegistryStore); !ok {
		t.Fatalf("NewStore() = %T, want RegistryStore", store)
	}
}
