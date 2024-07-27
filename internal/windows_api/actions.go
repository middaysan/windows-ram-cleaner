package windowsapi

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

const (
    SystemMemoryListInformationClass = 0x0050
    MemoryPurgeStandbyList           = 4
)

// Win32_PerfRawData_PerfOS_Memory structure for WMI queries
type Win32_PerfRawData_PerfOS_Memory struct {
    StandbyCacheNormalPriorityBytes uint64
    StandbyCacheReserveBytes        uint64
    StandbyCacheCoreBytes           uint64
    AvailableBytes                  uint64
}

// CleanOptions defines options for the CleanRAM function
type CleanOptions struct {
	IgnoreCritical bool
}

// DefaultCleanOptions returns a CleanOptions struct with default values
func DefaultCleanOptions() CleanOptions {
	return CleanOptions{
		IgnoreCritical: false,
	}
}

// CleanRAM cleans the system and process memory to free up RAM.
// It calls various functions to clean the system memory, process memory, and system working set.
// If any of the cleaning operations fail, it returns an error.
func CleanRAM(opts ...CleanOptions) error {
    var options CleanOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultCleanOptions()
	}

    if err := cleanSystemMemory(options.IgnoreCritical); err != nil {
        return fmt.Errorf("failed to clean system memory: %v", err)
    }

    if err := cleanProcessMemory(); err != nil {
        return fmt.Errorf("failed to clean process memory: %v", err)
    }

    if err := cleanSystemWorkingSet(); err != nil {
        return fmt.Errorf("failed to clean system working set: %v", err)
    }

    return nil
}

// CleanStandbyList purges the standby list in the system memory.
// It grants necessary privileges to the process and calls the NtSetSystemInformation function
// to perform the memory purge operation.
// Returns an error if granting privileges or calling NtSetSystemInformation fails.
func CleanStandbyList() error {
    if err := GrantPrivileges(); err != nil {
        return fmt.Errorf("failed to grant privileges to the process: %v", err)
    }

    memoryPurgeStandbyList := uint32(MemoryPurgeStandbyList)
    r1, _, err := NtSetSystemInformation.Call(
        uintptr(SystemMemoryListInformationClass),
        uintptr(unsafe.Pointer(&memoryPurgeStandbyList)),
        unsafe.Sizeof(memoryPurgeStandbyList),
    )
    if r1 != 0 {
        return fmt.Errorf("NtSetSystemInformation call failed: %v", err)
    }

    return nil
}

// GetStanbyListAndFreeRAMSize retrieves the size of the standby list and free memory using WMI
func GetStanbyListAndFreeRAMSize() (standbySize, freeSize uint64, err error) {
    var dst []Win32_PerfRawData_PerfOS_Memory
    query := wmi.CreateQuery(&dst, "")
    err = wmi.Query(query, &dst)
    if err != nil {
        return 0, 0, err
    }

    if len(dst) > 0 {
        standbySize = dst[0].StandbyCacheCoreBytes + dst[0].StandbyCacheNormalPriorityBytes + dst[0].StandbyCacheReserveBytes
        freeSize = dst[0].AvailableBytes
        return standbySize, freeSize, nil
    }
    return 0, 0, fmt.Errorf("no data returned from WMI query")
}

// getCurrentProcessHandle retrieves the handle of the current process
func getCurrentProcessHandle() windows.Handle {
    return windows.CurrentProcess()
}

// cleanProcessMemory sets the process working set size to minimum and maximum values
func cleanProcessMemory() error {
    hProcess := getCurrentProcessHandle()
    ret, _, err := ProcSetProcessWorkingSetSize.Call(uintptr(hProcess), uintptr(^uint32(0)), uintptr(^uint32(0)))
    if ret == 0 {
        return fmt.Errorf("failed to set process working set size: %v", err)
    }

    return nil
}

// cleanSystemWorkingSet empties the working set of the current process
func cleanSystemWorkingSet() error {
    hProcess := getCurrentProcessHandle()

    ret, _, err := ProcEmptyWorkingSet.Call(uintptr(hProcess))
    if ret == 0 {
        return fmt.Errorf("failed to empty working set: %v", err)
    }

    return nil
}

// cleanSystemMemory frees memory of non-critical processes by emptying their working sets
func cleanSystemMemory(ignoreCritical bool) error {
    snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
    if err != nil {
        return fmt.Errorf("failed to create snapshot: %v", err)
    }
    defer windows.CloseHandle(snapshot)

    var pe windows.ProcessEntry32
    pe.Size = uint32(unsafe.Sizeof(pe))
    if err := windows.Process32First(snapshot, &pe); err != nil {
        return fmt.Errorf("failed to get first process: %v", err)
    }

    for {
        if !ignoreCritical && isCriticalProcess(pe) {
            err = windows.Process32Next(snapshot, &pe)
            if err != nil {
                break
            }
            continue
        }

        hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_SET_QUOTA, false, pe.ProcessID)
        if err == nil {
            ProcEmptyWorkingSet.Call(uintptr(hProcess))
            windows.CloseHandle(hProcess)

            time.Sleep(10 * time.Millisecond)
        }

        err = windows.Process32Next(snapshot, &pe)
        if err != nil {
            break
        }
    }

    return nil
}

// isCriticalProcess checks if taskbar is visible
func IsTaskbarVisible() bool {
    taskbarHandle, _, _ := ProcFindWindowW.Call(
        uintptr(unsafe.Pointer(utf16PtrFromString("Shell_TrayWnd"))),
        uintptr(0),
    )
    visible, _, _ := ProcIsWindowVisible.Call(taskbarHandle)
    return visible != 0
}

func utf16PtrFromString(s string) *uint16 {
    ptr, err := syscall.UTF16PtrFromString(s)
    if err != nil {
        panic(err)
    }
    return ptr
}