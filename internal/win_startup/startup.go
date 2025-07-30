package winstartup

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

var WinTaskName = "WindowsRAMCleaner"

func CreateStartupTask() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {

		}
	}(key)

	err = key.SetStringValue(WinTaskName, exePath)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %v", err)
	}

	return nil
}

func DeleteStartupTask() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			fmt.Printf("failed to close registry key: %v\n", err)
		}
	}(key)

	err = key.DeleteValue(WinTaskName)
	if err != nil {
		return fmt.Errorf("failed to delete registry value: %v", err)
	}

	return nil
}

func IsStartupTaskExists() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("failed to open registry key: %v", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			fmt.Printf("failed to close registry key: %v\n", err)
		}
	}(key)

	_, valType, err := key.GetStringValue(WinTaskName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get registry value: %v", err)
	}

	return valType != 0, nil
}
