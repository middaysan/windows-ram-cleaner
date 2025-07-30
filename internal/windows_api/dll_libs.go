package windowsapi

import (
	"golang.org/x/sys/windows"
	"syscall"
)

// Windows API function definitions
var (
	// ModKernel32 DLLs
	ModKernel32              = syscall.NewLazyDLL("kernel32.dll")
	ModPSApi                 = syscall.NewLazyDLL("psapi.dll")
	Ntdll                    = syscall.NewLazyDLL("ntdll.dll")
	User32                   = syscall.NewLazyDLL("user32.dll")
	modKernel32              = syscall.NewLazyDLL("kernel32.dll")
	NtQuerySystemInformation = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtQuerySystemInformation")

	// ProcSetProcessWorkingSetSize Process functions
	ProcSetProcessWorkingSetSize = ModKernel32.NewProc("SetProcessWorkingSetSize")
	ProcEmptyWorkingSet          = ModPSApi.NewProc("EmptyWorkingSet")
	NtSetSystemInformation       = Ntdll.NewProc("NtSetSystemInformation")
	ProcMessageBoxW              = User32.NewProc("MessageBoxW")
	ProcFindWindowW              = User32.NewProc("FindWindowW")
	ProcIsWindowVisible          = User32.NewProc("IsWindowVisible")
	procGlobalMemoryStatusEx     = modKernel32.NewProc("GlobalMemoryStatusEx")
)
