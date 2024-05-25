// internal/memory/memory.go

package memory

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"clean-standby-list/internal/elevation"
	"clean-standby-list/internal/tray"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

const (
	SystemMemoryListInformationClass = 0x0050
	MemoryPurgeStandbyList           = 4
	PercentThreshold                 = 65
)

var lastCleanup time.Time

type Win32_PerfRawData_PerfOS_Memory struct {
	StandbyCacheNormalPriorityBytes uint64
	StandbyCacheReserveBytes        uint64
	StandbyCacheCoreBytes           uint64
	AvailableBytes                  uint64
}

// GetStandbyListInfo retrieves the size of the standby list and free memory using WMI
func GetStandbyListInfo() (standbySize, freeSize uint64, err error) {
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

func CheckAndCleanStandbyList() {
	standbySize, freeSize, err := GetStandbyListInfo()
	if err != nil {
		log.Printf("Error getting memory info: %v\n", err)
		return
	}
	percent := (standbySize * 100) / freeSize
	log.Printf("Standby List: %d MB, Free Memory: %d MB, Percent: %d%%\n", standbySize/1024/1024, freeSize/1024/1024, percent)

	if percent > PercentThreshold && time.Since(lastCleanup) > 5*time.Minute {
		log.Printf("Standby list exceeds %d%% of free memory, cleaning...\n", PercentThreshold)
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
