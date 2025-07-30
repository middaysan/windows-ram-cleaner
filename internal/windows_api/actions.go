package windowsapi

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	SystemMemoryListInformationClass = 0x50
	MemoryPurgeStandbyList           = 4
)

// MEMORYSTATUSEX for GlobalMemoryStatusEx
type MEMORYSTATUSEX struct {
	DwLength                uint32
	DwMemoryLoad            uint32
	UllTotalPhys            uint64
	UllAvailPhys            uint64
	UllTotalPageFile        uint64
	UllAvailPageFile        uint64
	UllTotalVirtual         uint64
	UllAvailVirtual         uint64
	UllAvailExtendedVirtual uint64
}

// SystemMemoryListInformation struct for NtQuerySystemInformation
type SystemMemoryListInformation struct {
	ZeroPageCount               uint64
	FreePageCount               uint64
	ModifiedPageCount           uint64
	ModifiedNoWritePageCount    uint64
	BadPageCount                uint64
	ActivePageCount             uint64
	StandbyPageCount            uint64
	StandbyPageCountLowPriority uint64
	StandbyPageCountNormal      uint64
	StandbyPageCountReserve     uint64
	TransitionPageCount         uint64
	ModifiedPageCountPagefile   uint64
}

// MemoryInfo represents memory stats
type MemoryInfo struct {
	FreeSize    uint64 // Available physical memory in bytes
	StandbySize uint64 // Standby cache size in bytes
}

// GetMemoryInfo gets memory info via WinAPI
func GetMemoryInfo() (MemoryInfo, error) {
	var memInfo MemoryInfo

	// 1. Free RAM
	var status MEMORYSTATUSEX
	status.DwLength = uint32(unsafe.Sizeof(status))
	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		return memInfo, fmt.Errorf("GlobalMemoryStatusEx failed: %v", err)
	}
	memInfo.FreeSize = status.UllAvailPhys

	// 2. Standby RAM
	bufferSize := uintptr(16 * 1024) // 16 KB
	buffer := make([]byte, bufferSize)

	ret, _, _ = NtQuerySystemInformation.Call(
		uintptr(SystemMemoryListInformationClass),
		uintptr(unsafe.Pointer(&buffer[0])),
		bufferSize,
		0,
	)

	if ret != 0 {
		return memInfo, fmt.Errorf("NtQuerySystemInformation failed: NTSTATUS=0x%x", ret)
	}

	pageSize := uint64(4096)
	counts := (*[128]uint64)(unsafe.Pointer(&buffer[0])) // берём большой массив, чтобы не выйти за границу

	// Standby страницы примерно с 6 по 15 индекс (включая high/low)
	standby := uint64(0)
	for i := 6; i <= 15; i++ {
		standby += counts[i]
	}

	memInfo.StandbySize = standby * pageSize
	return memInfo, nil
}

// CleanOptions for cleaning RAM
type CleanOptions struct {
	IgnoreCritical bool
}

// DefaultCleanOptions returns default CleanOptions
func DefaultCleanOptions() CleanOptions {
	return CleanOptions{
		IgnoreCritical: false,
	}
}

// CleanRAM cleans RAM: standby list, process WS, system WS
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

// CleanStandbyList purges standby list
func CleanStandbyList() error {
	if err := GrantPrivileges(); err != nil {
		return fmt.Errorf("failed to grant privileges: %v", err)
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

// cleanProcessMemory sets working set size to min/max
func cleanProcessMemory() error {
	hProcess := windows.CurrentProcess()
	ret, _, err := ProcSetProcessWorkingSetSize.Call(uintptr(hProcess), uintptr(^uint32(0)), uintptr(^uint32(0)))
	if ret == 0 {
		return fmt.Errorf("failed to set process working set size: %v", err)
	}
	return nil
}

// cleanSystemWorkingSet empties working set
func cleanSystemWorkingSet() error {
	hProcess := windows.CurrentProcess()
	ret, _, err := ProcEmptyWorkingSet.Call(uintptr(hProcess))
	if ret == 0 {
		return fmt.Errorf("failed to empty working set: %v", err)
	}
	return nil
}

// cleanSystemMemory frees memory of non-critical processes
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
		// Если процесс критический — пропускаем
		if ignoreCritical || !isCriticalProcess(pe) {
			hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_SET_QUOTA, false, pe.ProcessID)
			if err == nil {
				ret, _, err := ProcEmptyWorkingSet.Call(uintptr(hProcess))
				if ret == 0 {
					return err
				}
				err = windows.CloseHandle(hProcess)
				if err != nil {
					return err
				}
				time.Sleep(10 * time.Millisecond)
			}
		}

		if err := windows.Process32Next(snapshot, &pe); err != nil {
			break
		}
	}

	return nil
}

// IsTaskbarVisible checks taskbar visibility
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
