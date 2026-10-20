package internal

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// registryValue is the part of registry.Key that PATH reads and writes use.
// Tests replace the opener below with a fake, so the registry layer runs on any
// machine without touching the live hive.
type registryValue interface {
	GetStringValue(name string) (string, uint32, error)
	SetExpandStringValue(name, value string) error
	Close() error
}

// openRegistryKey is the single seam to the Windows registry API.
var openRegistryKey = func(root registry.Key, path string, access uint32) (registryValue, error) {
	return registry.OpenKey(root, path, access)
}

// getRegistryKeyAndName returns the registry root, subkey and value name that
// store the PATH for a scope.
func getRegistryKeyAndName(scope Scope) (registry.Key, string, string) {
	if scope == ScopeUser {
		return registry.CURRENT_USER, `Environment`, `PATH`
	}
	return registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, `Path`
}

// ReadPATH returns the PATH entries stored in the registry for a scope.
func ReadPATH(scope Scope) ([]string, error) {
	value, err := ReadRaw(scope)
	if err != nil {
		return nil, err
	}
	return SplitList(value), nil
}

// ReadRaw returns the unmodified registry value for a scope. It is used for
// backups and by ReadPATH.
func ReadRaw(scope Scope) (string, error) {
	root, path, valueName := getRegistryKeyAndName(scope)
	return readRaw(root, path, valueName)
}

func readRaw(root registry.Key, path, valueName string) (string, error) {
	k, err := openRegistryKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() { _ = k.Close() }()

	value, _, err := k.GetStringValue(valueName)
	if err != nil {
		if err == registry.ErrNotExist {
			return "", nil
		}
		return "", err
	}

	return value, nil
}

// WritePATH stores entries as the expanded-string PATH value for a scope.
func WritePATH(scope Scope, entries []string) error {
	root, path, valueName := getRegistryKeyAndName(scope)
	return writePATH(root, path, valueName, entries)
}

func writePATH(root registry.Key, path, valueName string, entries []string) error {
	k, err := openRegistryKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = k.Close() }()

	return k.SetExpandStringValue(valueName, JoinList(entries))
}

// BroadcastChange tells open applications that environment variables changed.
func BroadcastChange() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeoutW := user32.NewProc("SendMessageTimeoutW")

	const hwndBroadcast = 0xffff
	const wmSettingChange = 0x001A
	const smtoAbortIfHung = 0x0002

	envStr, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}

	_, _, _ = sendMessageTimeoutW.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(envStr)),
		uintptr(smtoAbortIfHung),
		uintptr(5000),
		0,
	)
}

// ensure the real key type still satisfies the seam.
var _ registryValue = registry.Key(0)
