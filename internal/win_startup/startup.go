package winstartup

import (
	"fmt"
	"os"
	"os/exec"
)

var WinTaskName = "WindowsRAMCleaner"

func CreateStartupTask() error {
    exePath, err := os.Executable()
    if err != nil {
        return err
    }

	// it fixes ui the bug that occur when the task runs before the system is ready
    delay := "0000:10"

    cmd := exec.Command("schtasks", "/create", "/tn", WinTaskName, "/tr", exePath, "/sc", "onlogon", "/rl", "highest", "/f", "/delay", delay)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to create scheduled task: %v, output: %s", err, string(output))
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
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to query scheduled task: %v, output: %s", err, string(output))
	}

	return true, nil
}
