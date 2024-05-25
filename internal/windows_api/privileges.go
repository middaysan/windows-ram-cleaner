// Description: This package provides functions to elevate the current process to run as administrator.
package windowsapi

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// RunAsAdmin runs the current executable with administrative privileges.
// It uses the Windows API function ShellExecute to execute the executable
// with the "runas" verb, which prompts the user for consent to elevate
// the process. The function takes no arguments and returns no values.
// If an error occurs during the execution, it will be printed to the console.
func RunAsAdmin() {
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

// GrantPrivileges grants specific privileges to the current process.
// It enables the privileges specified in the `privileges` slice.
// If any privilege fails to be enabled, an error is returned.
func GrantPrivileges() error {
	privileges := []string{"SeProfileSingleProcessPrivilege", "SeIncreaseQuotaPrivilege", "SeDebugPrivilege"}
	for _, privilege := range privileges {
		if err := enablePrivilege(privilege); err != nil {
			return fmt.Errorf("failed to enable privilege %s: %v", privilege, err)
		}
	}

	return nil
}

// enablePrivilege enables the specified privilege for the current process.
// It takes a privilegeName string as input and returns an error if any.
// The function uses the Windows API to lookup the privilege value, open the process token,
// adjust the token privileges, and enable the specified privilege.
// If any error occurs during the process, it returns an error with a descriptive message.
func enablePrivilege(privilegeName string) error {
	var luid windows.LUID
	err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(privilegeName), &luid)
	if err != nil {
		return fmt.Errorf("LookupPrivilegeValue error: %v", err)
	}

	var token windows.Token
	processHandle := windows.CurrentProcess()

	err = windows.OpenProcessToken(processHandle, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("OpenProcessToken error: %v", err)
	}
	defer token.Close()

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

	return nil
}

func IsRunAsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
