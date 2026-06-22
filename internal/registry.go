package internal

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func getRegistryKeyAndName(scope string) (registry.Key, string, string) {
	if scope == "User" {
		return registry.CURRENT_USER, `Environment`, `PATH`
	}
	return registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, `Path`
}

func ReadPATH(scope string) ([]string, error) {
	rootKey, path, valueName := getRegistryKeyAndName(scope)

	k, err := registry.OpenKey(rootKey, path, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	val, _, err := k.GetStringValue(valueName)
	if err != nil {
		if err == registry.ErrNotExist {
			return []string{}, nil
		}
		return nil, err
	}

	parts := strings.Split(val, ";")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result, nil
}

func WritePATH(scope string, entries []string) error {
	rootKey, path, valueName := getRegistryKeyAndName(scope)

	k, err := registry.OpenKey(rootKey, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	val := strings.Join(entries, ";")
	err = k.SetExpandStringValue(valueName, val)
	if err != nil {
		return err
	}

	return nil
}

func BroadcastChange() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeoutW := user32.NewProc("SendMessageTimeoutW")

	const HWND_BROADCAST = 0xffff
	const WM_SETTINGCHANGE = 0x001A
	const SMTO_ABORTIFHUNG = 0x0002

	envStr, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}

	sendMessageTimeoutW.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(envStr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000),
		0,
	)
}
