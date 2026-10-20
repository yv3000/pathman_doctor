package internal

import (
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestGetRegistryKeyAndName(t *testing.T) {
	tests := []struct {
		name      string
		scope     Scope
		wantRoot  registry.Key
		wantPath  string
		wantValue string
	}{
		{
			name:      "user",
			scope:     ScopeUser,
			wantRoot:  registry.CURRENT_USER,
			wantPath:  `Environment`,
			wantValue: `PATH`,
		},
		{
			name:      "system",
			scope:     ScopeSystem,
			wantRoot:  registry.LOCAL_MACHINE,
			wantPath:  `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
			wantValue: `Path`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, path, value := getRegistryKeyAndName(tt.scope)
			if root != tt.wantRoot {
				t.Fatalf("root = %v, want %v", root, tt.wantRoot)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q, want %q", path, tt.wantPath)
			}
			if value != tt.wantValue {
				t.Fatalf("value name = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

type fakeRegistryKey struct {
	value    string
	valueErr error
	setErr   error
	setName  string
	setValue string
	closed   bool
}

func (f *fakeRegistryKey) GetStringValue(string) (string, uint32, error) {
	if f.valueErr != nil {
		return "", 0, f.valueErr
	}
	return f.value, uint32(registry.SZ), nil
}

func (f *fakeRegistryKey) SetExpandStringValue(name, value string) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.setName, f.setValue = name, value
	return nil
}

func (f *fakeRegistryKey) Close() error {
	f.closed = true
	return nil
}

type fakeOpener struct {
	key    *fakeRegistryKey
	err    error
	root   registry.Key
	path   string
	access uint32
	calls  int
}

func (o *fakeOpener) open(root registry.Key, path string, access uint32) (registryValue, error) {
	o.calls++
	o.root, o.path, o.access = root, path, access
	if o.err != nil {
		return nil, o.err
	}
	return o.key, nil
}

func installFakeOpener(t *testing.T, opener *fakeOpener) {
	t.Helper()
	previous := openRegistryKey
	openRegistryKey = opener.open
	t.Cleanup(func() { openRegistryKey = previous })
}

func TestReadRawReturnsStoredValue(t *testing.T) {
	key := &fakeRegistryKey{value: `C:\Windows;%USERPROFILE%\bin`}
	opener := &fakeOpener{key: key}
	installFakeOpener(t, opener)

	got, err := ReadRaw(ScopeUser)
	if err != nil {
		t.Fatalf("ReadRaw(User) = %v", err)
	}
	if got != key.value {
		t.Fatalf("ReadRaw(User) = %q, want %q", got, key.value)
	}
	if opener.path != `Environment` {
		t.Fatalf("opened %q, want Environment", opener.path)
	}
	if opener.access != registry.QUERY_VALUE {
		t.Fatalf("access = %d, want QUERY_VALUE", opener.access)
	}
	if !key.closed {
		t.Fatal("registry key was not closed")
	}
}

func TestReadRawTreatsMissingValueAsEmpty(t *testing.T) {
	key := &fakeRegistryKey{valueErr: registry.ErrNotExist}
	installFakeOpener(t, &fakeOpener{key: key})

	got, err := ReadRaw(ScopeUser)
	if err != nil {
		t.Fatalf("ReadRaw(User) = %v, want no error for a missing value", err)
	}
	if got != "" {
		t.Fatalf("ReadRaw(User) = %q, want empty", got)
	}
}

func TestReadRawPropagatesOpenError(t *testing.T) {
	opener := &fakeOpener{err: errors.New("access denied")}
	installFakeOpener(t, opener)

	if _, err := ReadRaw(ScopeSystem); err == nil {
		t.Fatal("ReadRaw(System) should surface an open error")
	}
	if opener.access != registry.QUERY_VALUE {
		t.Fatalf("access = %d, want QUERY_VALUE", opener.access)
	}
}

func TestReadRawPropagatesGetError(t *testing.T) {
	key := &fakeRegistryKey{valueErr: errors.New("unexpected type")}
	installFakeOpener(t, &fakeOpener{key: key})

	if _, err := ReadRaw(ScopeUser); err == nil {
		t.Fatal("ReadRaw(User) should surface a GetStringValue error")
	}
	if !key.closed {
		t.Fatal("registry key was not closed after a read error")
	}
}

func TestReadPATHSplitsStoredValue(t *testing.T) {
	key := &fakeRegistryKey{value: `C:\one;;  C:\two ;`}
	installFakeOpener(t, &fakeOpener{key: key})

	got, err := ReadPATH(ScopeUser)
	if err != nil {
		t.Fatalf("ReadPATH(User) = %v", err)
	}
	if len(got) != 2 || got[0] != `C:\one` || got[1] != `C:\two` {
		t.Fatalf("ReadPATH(User) = %#v, want [C:\\\\one C:\\\\two]", got)
	}
}

func TestReadPATHPropagatesReadError(t *testing.T) {
	installFakeOpener(t, &fakeOpener{err: errors.New("access denied")})

	if _, err := ReadPATH(ScopeUser); err == nil {
		t.Fatal("ReadPATH(User) should surface the read error")
	}
}

func TestWritePATHJoinsEntriesForScope(t *testing.T) {
	key := &fakeRegistryKey{}
	opener := &fakeOpener{key: key}
	installFakeOpener(t, opener)

	err := WritePATH(ScopeSystem, []string{`C:\Windows`, `C:\Tools`})
	if err != nil {
		t.Fatalf("WritePATH(System) = %v", err)
	}
	if opener.path != `SYSTEM\CurrentControlSet\Control\Session Manager\Environment` {
		t.Fatalf("opened %q, want the System Environment key", opener.path)
	}
	if opener.access != registry.SET_VALUE {
		t.Fatalf("access = %d, want SET_VALUE", opener.access)
	}
	if key.setName != `Path` {
		t.Fatalf("value name = %q, want Path", key.setName)
	}
	if want := `C:\Windows;C:\Tools`; key.setValue != want {
		t.Fatalf("written value = %q, want %q", key.setValue, want)
	}
	if !key.closed {
		t.Fatal("registry key was not closed")
	}
}

func TestWritePATHPropagatesOpenError(t *testing.T) {
	key := &fakeRegistryKey{}
	installFakeOpener(t, &fakeOpener{key: key, err: errors.New("access denied")})

	if err := WritePATH(ScopeSystem, []string{`C:\Windows`}); err == nil {
		t.Fatal("WritePATH(System) should surface an open error")
	}
	if key.setValue != "" {
		t.Fatalf("value was written despite the open error: %q", key.setValue)
	}
}

func TestWritePATHPropagatesSetError(t *testing.T) {
	key := &fakeRegistryKey{setErr: errors.New("access denied")}
	installFakeOpener(t, &fakeOpener{key: key})

	if err := WritePATH(ScopeUser, []string{`C:\Windows`}); err == nil {
		t.Fatal("WritePATH(User) should surface a SetExpandStringValue error")
	}
	if !key.closed {
		t.Fatal("registry key was not closed after a write error")
	}
}

// TestReadPATHUserIsReadable reads the live HKCU hive, so it is opt-in. The
// default suite is hermetic: registry reads and writes run against the fake
// key above.
func TestReadPATHUserIsReadable(t *testing.T) {
	if os.Getenv("PATHMAN_LIVE_REGISTRY") != "1" {
		t.Skip("live registry check: set PATHMAN_LIVE_REGISTRY=1 to run it")
	}

	entries, err := ReadPATH(ScopeUser)
	if err != nil {
		t.Fatalf("ReadPATH(User) = %v", err)
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry) == "" {
			t.Fatalf("ReadPATH(User) returned an empty entry: %#v", entries)
		}
	}
}
