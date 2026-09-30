//go:build windows

package config

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func systemTimezone() (string, error) {
	keyName, err := windowsTimezoneKeyName()
	if err != nil {
		return "", err
	}
	if name, ok := windowsTimezoneToIANA[keyName]; ok {
		return name, nil
	}
	return keyName, nil
}

func nativeTimezoneLocation(name string) (*time.Location, bool) {
	keyName, err := windowsTimezoneKeyName()
	if err != nil || name != keyName {
		return nil, false
	}
	return time.Local, true
}

func windowsTimezoneKeyName() (string, error) {
	path, err := syscall.UTF16PtrFromString(`SYSTEM\CurrentControlSet\Control\TimeZoneInformation`)
	if err != nil {
		return "", err
	}
	var key syscall.Handle
	if err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, path, 0, syscall.KEY_READ, &key); err != nil {
		return "", fmt.Errorf("read Windows time zone: %w", err)
	}
	defer syscall.RegCloseKey(key)

	valueName, err := syscall.UTF16PtrFromString("TimeZoneKeyName")
	if err != nil {
		return "", err
	}
	var valueType uint32
	var size uint32
	if err := syscall.RegQueryValueEx(key, valueName, nil, &valueType, nil, &size); err != nil {
		return "", fmt.Errorf("read Windows time zone name: %w", err)
	}
	if valueType != syscall.REG_SZ || size < 2 {
		return "", fmt.Errorf("Windows time zone registry value has an unexpected format")
	}
	buffer := make([]byte, size)
	if err := syscall.RegQueryValueEx(key, valueName, nil, &valueType, &buffer[0], &size); err != nil {
		return "", fmt.Errorf("read Windows time zone name: %w", err)
	}
	units := unsafe.Slice((*uint16)(unsafe.Pointer(&buffer[0])), len(buffer)/2)
	return strings.TrimSpace(syscall.UTF16ToString(units)), nil
}

var windowsTimezoneToIANA = map[string]string{
	"UTC":                            "Etc/UTC",
	"Dateline Standard Time":         "Etc/GMT+12",
	"UTC-11":                         "Pacific/Pago_Pago",
	"Aleutian Standard Time":         "America/Adak",
	"Hawaiian Standard Time":         "Pacific/Honolulu",
	"Alaskan Standard Time":          "America/Anchorage",
	"Pacific Standard Time":          "America/Los_Angeles",
	"Pacific Standard Time (Mexico)": "America/Tijuana",
	"US Mountain Standard Time":      "America/Phoenix",
	"Mountain Standard Time":         "America/Denver",
	"Central Standard Time":          "America/Chicago",
	"Central Standard Time (Mexico)": "America/Mexico_City",
	"Canada Central Standard Time":   "America/Regina",
	"Eastern Standard Time":          "America/New_York",
	"US Eastern Standard Time":       "America/Indianapolis",
	"Atlantic Standard Time":         "America/Halifax",
	"Newfoundland Standard Time":     "America/St_Johns",
	"SA Pacific Standard Time":       "America/Bogota",
	"Argentina Standard Time":        "America/Buenos_Aires",
	"E. South America Standard Time": "America/Sao_Paulo",
	"Greenland Standard Time":        "America/Nuuk",
	"Azores Standard Time":           "Atlantic/Azores",
	"GMT Standard Time":              "Europe/London",
	"W. Europe Standard Time":        "Europe/Berlin",
	"Central Europe Standard Time":   "Europe/Budapest",
	"Romance Standard Time":          "Europe/Paris",
	"Turkey Standard Time":           "Europe/Istanbul",
	"Russian Standard Time":          "Europe/Moscow",
	"Israel Standard Time":           "Asia/Jerusalem",
	"Arabian Standard Time":          "Asia/Dubai",
	"Iran Standard Time":             "Asia/Tehran",
	"Afghanistan Standard Time":      "Asia/Kabul",
	"Pakistan Standard Time":         "Asia/Karachi",
	"India Standard Time":            "Asia/Kolkata",
	"Nepal Standard Time":            "Asia/Kathmandu",
	"Bangladesh Standard Time":       "Asia/Dhaka",
	"SE Asia Standard Time":          "Asia/Bangkok",
	"China Standard Time":            "Asia/Shanghai",
	"Singapore Standard Time":        "Asia/Singapore",
	"Taipei Standard Time":           "Asia/Taipei",
	"Tokyo Standard Time":            "Asia/Tokyo",
	"Korea Standard Time":            "Asia/Seoul",
	"Yakutsk Standard Time":          "Asia/Yakutsk",
	"AUS Central Standard Time":      "Australia/Darwin",
	"Cen. Australia Standard Time":   "Australia/Adelaide",
	"AUS Eastern Standard Time":      "Australia/Sydney",
	"E. Australia Standard Time":     "Australia/Brisbane",
	"Tasmania Standard Time":         "Australia/Hobart",
	"New Zealand Standard Time":      "Pacific/Auckland",
	"Fiji Standard Time":             "Pacific/Fiji",
	"Tonga Standard Time":            "Pacific/Tongatapu",
}
