// cmd/main.go

package main

import (
	"runtime"
	"time"
	"windows-ram-cleaner/internal/tray"
	"windows-ram-cleaner/internal/windows_api"

	_ "github.com/josephspurrier/goversioninfo"

	"github.com/getlantern/systray"
)

var stopChan = make(chan struct{})

//go:generate goversioninfo -icon=exe_icon.ico -manifest=app.manifest
func main() {
	// Request admin rights if not already granted
	if !windowsapi.IsRunAsAdmin() {
		windowsapi.RequestAdminRights()
		return
	}

	// Wait for the taskbar to be visible before running the system tray
	for !windowsapi.IsTaskbarVisible() {
		time.Sleep(1 * time.Second)
	}

	go autoUpdateTooltip(stopChan)

	systray.Run(tray.OnReady, onExit)
	tray.UpdateTooltip()
}

func onExit() {
	close(stopChan)
	runtime.GC()
}

// autoUpdateTooltip periodically updates the tooltip text of the system tray icon.
func autoUpdateTooltip(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		default:
			time.Sleep(2 * time.Second)
			tray.UpdateTooltip()
		}
	}
}
