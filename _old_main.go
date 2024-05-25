package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

const (
	SystemMemoryListInformationClass = 0x0050
	MemoryPurgeStandbyList           = 4
)

// Embed the icon using go:embed
//
//go:embed icon.ico
var iconData []byte

var lastCleanup time.Time

// Check and elevate to run as administrator
func runMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 // SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println(err)
	}
}

// enablePrivilege enables a specified privilege for the current process
func enablePrivilege(privilegeName string) error {
	log.Printf("Attempting to enable privilege: %s\n", privilegeName)
	var luid windows.LUID
	err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(privilegeName), &luid)
	if err != nil {
		return fmt.Errorf("LookupPrivilegeValue error: %v", err)
	}
	log.Printf("LookupPrivilegeValue succeeded for privilege: %s\n", privilegeName)

	var token windows.Token
	processHandle := windows.CurrentProcess()
	log.Println("GetCurrentProcess succeeded")

	err = windows.OpenProcessToken(processHandle, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("OpenProcessToken error: %v", err)
	}
	defer token.Close()
	log.Println("OpenProcessToken succeeded")

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED},
		},
	}

	err = windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil)
	if err != nil {
		return fmt.Errorf("AdjustTokenPrivileges error: %v", err)
	}
	log.Printf("AdjustTokenPrivileges succeeded for privilege: %s\n", privilegeName)
	return nil
}

// emptyStandbyList performs the cleanup of the standby list
func emptyStandbyList() error {
	privileges := []string{"SeProfileSingleProcessPrivilege", "SeIncreaseQuotaPrivilege", "SeDebugPrivilege"}
	for _, privilege := range privileges {
		if err := enablePrivilege(privilege); err != nil {
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

// getStandbyListInfo retrieves the size of the standby list and free memory
func getStandbyListInfo() (standbySize, freeSize uint64, err error) {
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

// checkAndCleanStandbyList checks the size of the standby list and cleans it if necessary
func checkAndCleanStandbyList() {
	standbySize, freeSize, err := getStandbyListInfo()
	if err != nil {
		log.Printf("Error getting memory info: %v\n", err)
		return
	}
	percent := (standbySize * 100) / freeSize
	log.Printf("Standby List: %d MB, Free Memory: %d MB, Percent: %d%%\n", standbySize/1024/1024, freeSize/1024/1024, percent)

	if percent > 65 && time.Since(lastCleanup) > 5*time.Minute {
		log.Println("Standby list exceeds 65% of free memory, cleaning...")
		if err := emptyStandbyList(); err != nil {
			log.Printf("Error cleaning standby list: %v\n", err)
		} else {
			log.Println("Standby list cleaned successfully")
			lastCleanup = time.Now()
		}
	}
	updateTooltip(standbySize, freeSize, percent)
}

// updateTooltip updates the system tray tooltip with the current standby list info
func updateTooltip(standbySize, freeSize uint64, percent uint64) {
	tooltip := fmt.Sprintf("Standby List: %d MB, Free Memory: %d MB, Percent: %d%%", standbySize/1024/1024, freeSize/1024/1024, percent)
	systray.SetTooltip(tooltip)
}

// onReady is called when the system tray is ready
func onReady() {
	systray.SetIcon(iconData) // Use the embedded icon
	systray.SetTitle("Memory Cleaner")
	systray.SetTooltip("Right-click to clean standby list")

	mClean := systray.AddMenuItem("Clean", "Clean the standby list")
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	go func() {
		for {
			select {
			case <-mClean.ClickedCh:
				if err := emptyStandbyList(); err != nil {
					log.Printf("Error cleaning standby list: %v\n", err)
				} else {
					log.Println("Standby list cleaned successfully")
					lastCleanup = time.Now()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()

	// Periodically update the tooltip with memory information
	go func() {
		for {
			checkAndCleanStandbyList()
			time.Sleep(1 * time.Minute) // Update every minute
		}
	}()
}

func onExit() {
	runtime.GC()
}

// Build with: go build -ldflags="-H windowsgui -extldflags=-Wl,app.manifest"
func main() {
	// Request admin rights if not already granted
	if !isRunAsAdmin() {
		runMeElevated()
		return
	}

	systray.Run(onReady, onExit)
}

// isRunAsAdmin checks if the application is running with admin rights
func isRunAsAdmin() bool {
	elevated := windows.GetCurrentProcessToken().IsElevated()
	fmt.Printf("Admin: %v\n", elevated)
	return elevated
}
