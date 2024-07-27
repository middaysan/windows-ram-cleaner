package windowsapi

import "syscall"

// Windows API function definitions
var (
	// DLLs
	ModKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	ModPsapi                     = syscall.NewLazyDLL("psapi.dll")
	Ntdll                        = syscall.NewLazyDLL("ntdll.dll")
	User32          			 = syscall.NewLazyDLL("user32.dll")
    
	// Process functions
	ProcSetProcessWorkingSetSize = ModKernel32.NewProc("SetProcessWorkingSetSize")
	ProcEmptyWorkingSet          = ModPsapi.NewProc("EmptyWorkingSet")
	NtSetSystemInformation       = Ntdll.NewProc("NtSetSystemInformation")
	ProcMessageBoxW 			 = User32.NewProc("MessageBoxW")
	ProcFindWindowW    			 = User32.NewProc("FindWindowW")
    ProcIsWindowVisible 		 = User32.NewProc("IsWindowVisible")
)
