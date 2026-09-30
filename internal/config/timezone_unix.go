//go:build !windows

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func systemTimezone() (string, error) {
	if value, ok := os.LookupEnv("TZ"); ok && value != "" {
		name := strings.TrimPrefix(value, ":")
		if filepath.IsAbs(name) {
			name = zoneNameFromPath(name)
		}
		if isLoadableTimezone(name) {
			return name, nil
		}
	}

	if path, err := os.Readlink("/etc/localtime"); err == nil {
		if name := zoneNameFromPath(path); isLoadableTimezone(name) {
			return name, nil
		}
	}

	name := time.Local.String()
	if name != "" && name != "Local" && isLoadableTimezone(name) {
		return name, nil
	}
	return "", fmt.Errorf("could not determine the operating system time zone as an IANA name")
}

func nativeTimezoneLocation(string) (*time.Location, bool) {
	return nil, false
}

func zoneNameFromPath(path string) string {
	path = filepath.ToSlash(path)
	const marker = "/zoneinfo/"
	if index := strings.LastIndex(path, marker); index >= 0 {
		return strings.TrimPrefix(path[index+len(marker):], "posix/")
	}
	return ""
}

func isLoadableTimezone(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}
