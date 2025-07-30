package windowsapi

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestIsCriticalProcess(t *testing.T) {
	tests := []struct {
		name     string
		exeFile  string
		expected bool
	}{
		{
			name:     "critical process",
			exeFile:  "explorer.exe",
			expected: true,
		},
		{
			name:     "another critical process",
			exeFile:  "lsass.exe",
			expected: true,
		},
		{
			name:     "non-critical process",
			exeFile:  "notepad.exe",
			expected: false,
		},
		{
			name:     "empty process name",
			exeFile:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ProcessEntry32 with the test exe file name
			var pe windows.ProcessEntry32
			copy(pe.ExeFile[:], windows.StringToUTF16(tt.exeFile))

			// Call the function
			result := isCriticalProcess(pe)

			// Check the result
			if result != tt.expected {
				t.Errorf("isCriticalProcess() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCriticalProcessesMap(t *testing.T) {
	// Test that some expected critical processes are in the map
	expectedCritical := []string{
		"csrss.exe",
		"wininit.exe",
		"services.exe",
		"lsass.exe",
		"explorer.exe",
	}

	for _, procName := range expectedCritical {
		if _, exists := CriticalProcesses[procName]; !exists {
			t.Errorf("Expected %s to be in CriticalProcesses map", procName)
		}
	}

	// Test that the map doesn't contain some non-critical processes
	nonCritical := []string{
		"notepad.exe",
		"calc.exe",
		"mspaint.exe",
	}

	for _, procName := range nonCritical {
		if _, exists := CriticalProcesses[procName]; exists {
			t.Errorf("Did not expect %s to be in CriticalProcesses map", procName)
		}
	}
}
