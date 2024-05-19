package main

import (
	_ "embed" // Import the embed package
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
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
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
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
