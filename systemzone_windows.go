//go:build windows

package wantai

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows records the zone under its own name — "Tokyo Standard Time" — in the
// registry, and nothing in the standard library hands it over. time.Local knows
// the offset and an abbreviation, and the registry's StandardName is localised
// (a Japanese Windows writes Japanese there), so TimeZoneKeyName is the only
// value that is both present and stable. It has been there since Vista.
//
// Go ignores TZ on Windows — time.Local comes from GetTimeZoneInformation — so
// TZ is not consulted here either. Honouring it would make this disagree with
// the zone the process is actually rendering in.
const (
	tziKeyPath  = `SYSTEM\CurrentControlSet\Control\TimeZoneInformation`
	tziKeyValue = "TimeZoneKeyName"
)

func systemZoneName() (string, error) {
	win, err := readTimeZoneKeyName()
	if err != nil {
		return "", err
	}
	iana, ok := windowsZones[win]
	if !ok {
		// A zone Windows has and CLDR does not, or a table that has aged out.
		// Saying which name failed is the whole value of the error here.
		return "", fmt.Errorf("wantai: Windows reports the timezone %q, which is not in this build's CLDR mapping", win)
	}
	return iana, nil
}

// readTimeZoneKeyName reads the value out of the registry using only the
// standard library. golang.org/x/sys/windows/registry would be the usual way,
// but this package has no dependencies and syscall already exports what three
// registry calls need.
func readTimeZoneKeyName() (string, error) {
	path, err := syscall.UTF16PtrFromString(tziKeyPath)
	if err != nil {
		return "", err
	}
	var key syscall.Handle
	if err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, path, 0, syscall.KEY_READ, &key); err != nil {
		return "", fmt.Errorf("wantai: cannot open HKLM\\%s: %w", tziKeyPath, err)
	}
	defer syscall.RegCloseKey(key)

	name, err := syscall.UTF16PtrFromString(tziKeyValue)
	if err != nil {
		return "", err
	}

	// Ask for the size first: the value is a zone name, so it is short, but a
	// fixed buffer is the kind of guess that is wrong exactly once.
	var valueType, size uint32
	if err := syscall.RegQueryValueEx(key, name, nil, &valueType, nil, &size); err != nil {
		return "", fmt.Errorf("wantai: cannot read %s: %w", tziKeyValue, err)
	}
	if valueType != syscall.REG_SZ {
		return "", fmt.Errorf("wantai: %s is registry type %d, want REG_SZ", tziKeyValue, valueType)
	}
	buf := make([]byte, size)
	if err := syscall.RegQueryValueEx(key, name, nil, &valueType, &buf[0], &size); err != nil {
		return "", fmt.Errorf("wantai: cannot read %s: %w", tziKeyValue, err)
	}

	// The bytes are UTF-16; size is in bytes and includes the terminating NUL.
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(&buf[0])), len(buf)/2)
	return syscall.UTF16ToString(u16), nil
}
