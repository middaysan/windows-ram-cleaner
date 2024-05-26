package windowsapi

import (
    "syscall"
    "unsafe"
)

const (
    MB_OK               = 0x00000000
    MB_ICONERROR        = 0x00000010
    MB_ICONINFORMATION  = 0x00000040
    MB_ICONWARNING      = 0x00000030
    MB_ICONQUESTION     = 0x00000020
    MB_SYSTEMMODAL      = 0x00001000
)

func messageBox(hwnd uintptr, text, caption string, flags uint) int {
    t, _ := syscall.UTF16PtrFromString(text)
    c, _ := syscall.UTF16PtrFromString(caption)
    ret, _, _ := ProcMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), uintptr(flags))
    return int(ret)
}

func ShowInfo(text, caption string) {
    messageBox(0, text, caption, MB_OK|MB_ICONINFORMATION|MB_SYSTEMMODAL)
}

func ShowWarning(text, caption string) {
    messageBox(0, text, caption, MB_OK|MB_ICONWARNING|MB_SYSTEMMODAL)
}

func ShowError(text, caption string) {
    messageBox(0, text, caption, MB_OK|MB_ICONERROR|MB_SYSTEMMODAL)
}

func ShowQuestion(text, caption string) {
    messageBox(0, text, caption, MB_OK|MB_ICONQUESTION|MB_SYSTEMMODAL)
}
