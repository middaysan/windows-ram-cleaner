// Description: This package provides functions to elevate the current process to run as administrator.
package elevation

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func RunMeElevated() {
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

func EnablePrivilege(privilegeName string) error {
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

func IsRunAsAdmin() bool {
	elevated := windows.GetCurrentProcessToken().IsElevated()
	fmt.Printf("Admin: %v\n", elevated)
	return elevated
}
