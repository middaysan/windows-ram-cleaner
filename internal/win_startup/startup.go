package winstartup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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
	cmd := exec.Command("schtasks", "/delete", "/tn", WinTaskName, "/f")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to delete scheduled task: %v, output: %s", err, string(output))
	}

	return nil
}

func IsStartupTaskExists() (bool, error) {
	cmd := exec.Command("schtasks", "/query", "/tn", WinTaskName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to query scheduled task: %v, output: %s", err, string(output))
	}

	return true, nil
}
