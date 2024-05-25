// internal/memory/memory.go

package memory

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"clean-standby-list/internal/elevation"
	"clean-standby-list/internal/tray"

	"golang.org/x/sys/windows"
)

const (
	SystemMemoryListInformationClass = 0x0050
	MemoryPurgeStandbyList           = 4
)

var lastCleanup time.Time

func GetStandbyListInfo() (standbySize, freeSize uint64, err error) {
	// Define the MEMORYSTATUSEX structure
	type MEMORYSTATUSEX struct {
		Length               uint32
		MemoryLoad           uint32
		TotalPhys            uint64
		AvailPhys            uint64
		TotalPageFile        uint64
		AvailPageFile        uint64
		TotalVirtual         uint64
		AvailVirtual         uint64
		AvailExtendedVirtual uint64
	}

	// Initialize the structure and set its length
	var memStatus MEMORYSTATUSEX
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))

	// Call GlobalMemoryStatusEx to fill the structure
	r, _, err := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(uintptr(unsafe.Pointer(&memStatus)))
	if r == 0 {
		return 0, 0, err
	}

	// Calculate standby size and free size
	standbySize = memStatus.TotalPageFile - memStatus.AvailPageFile
	freeSize = memStatus.AvailPhys + standbySize

	return standbySize, freeSize, nil
}

func CheckAndCleanStandbyList() {
	standbySize, freeSize, err := GetStandbyListInfo()
	if err != nil {
		log.Printf("Error getting memory info: %v\n", err)
		return
	}
	percent := (standbySize * 100) / freeSize
	log.Printf("Standby List: %d MB, Free Memory: %d MB, Percent: %d%%\n", standbySize/1024/1024, freeSize/1024/1024, percent)

	if percent > 65 && time.Since(lastCleanup) > 5*time.Minute {
		log.Println("Standby list exceeds 65% of free memory, cleaning...")
		if err := EmptyStandbyList(); err != nil {
			log.Printf("Error cleaning standby list: %v\n", err)
		} else {
			log.Println("Standby list cleaned successfully")
			lastCleanup = time.Now()
		}
	}
	tray.UpdateTooltip(standbySize, freeSize, percent)
}

func EmptyStandbyList() error {
	privileges := []string{"SeProfileSingleProcessPrivilege", "SeIncreaseQuotaPrivilege", "SeDebugPrivilege"}
	for _, privilege := range privileges {
		if err := elevation.EnablePrivilege(privilege); err != nil {
			return fmt.Errorf("failed to enable privilege %s: %v", privilege, err)
		}
		log.Printf("Enabled privilege: %s\n", privilege)
	}
	log.Println("All privileges enabled successfully")

	ntdll := windows.NewLazySystemDLL("ntdll.dll")
	if err := ntdll.Load(); err != nil {
		return fmt.Errorf("failed to load ntdll.dll: %v", err)
	}
	log.Println("ntdll.dll loaded successfully")

	procNtSetSystemInformation := ntdll.NewProc("NtSetSystemInformation")
	memoryPurgeStandbyList := uint32(MemoryPurgeStandbyList)
	r1, _, err := procNtSetSystemInformation.Call(
		uintptr(SystemMemoryListInformationClass),
		uintptr(unsafe.Pointer(&memoryPurgeStandbyList)),
		unsafe.Sizeof(memoryPurgeStandbyList),
	)
	if r1 != 0 {
		return fmt.Errorf("NtSetSystemInformation call failed: %v", err)
	}
	log.Println("NtSetSystemInformation call succeeded")

	return nil
}
