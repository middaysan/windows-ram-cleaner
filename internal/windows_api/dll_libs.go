package windowsapi

import "syscall"

// Windows API function definitions
var (
	ModKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	ModPsapi                     = syscall.NewLazyDLL("psapi.dll")
	Ntdll                        = syscall.NewLazyDLL("ntdll.dll")
	ProcSetProcessWorkingSetSize = ModKernel32.NewProc("SetProcessWorkingSetSize")
	ProcEmptyWorkingSet          = ModPsapi.NewProc("EmptyWorkingSet")
	NtSetSystemInformation       = Ntdll.NewProc("NtSetSystemInformation")
)