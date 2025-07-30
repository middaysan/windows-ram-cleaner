package windowsapi

import (
	"syscall"
	"unsafe"
)

const (
	MbOk              = 0x00000000
	MbIconError       = 0x00000010
	MbIconInformation = 0x00000040
	MbIconWarning     = 0x00000030
	MbIconQuestion    = 0x00000020
	MbSystemModal     = 0x00001000
)

func messageBox(hwnd uintptr, text, caption string, flags uint) int {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(caption)
	ret, _, _ := ProcMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), uintptr(flags))
	return int(ret)
}

func ShowInfo(text, caption string) {
	messageBox(0, text, caption, MbOk|MbIconInformation|MbSystemModal)
}

func ShowWarning(text, caption string) {
	messageBox(0, text, caption, MbOk|MbIconWarning|MbSystemModal)
}

func ShowError(text, caption string) {
	messageBox(0, text, caption, MbOk|MbIconError|MbSystemModal)
}

func ShowQuestion(text, caption string) {
	messageBox(0, text, caption, MbOk|MbIconQuestion|MbSystemModal)
}
