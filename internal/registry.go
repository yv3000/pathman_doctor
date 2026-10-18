package internal

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

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
	rootKey, path, valueName := getRegistryKeyAndName(scope)

	k, err := registry.OpenKey(rootKey, path, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

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
	rootKey, path, valueName := getRegistryKeyAndName(scope)

	k, err := registry.OpenKey(rootKey, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.SetExpandStringValue(valueName, strings.Join(entries, ";"))
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
