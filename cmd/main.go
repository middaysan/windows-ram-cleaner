// cmd/main.go

package main

import (
	"clean-standby-list/internal/elevation"
	"clean-standby-list/internal/memory"
	"clean-standby-list/internal/tray"
	"runtime"

	"github.com/getlantern/systray"
)

func main() {
	// Request admin rights if not already granted
	if !elevation.IsRunAsAdmin() {
		elevation.RunMeElevated()
		return
	}

	onReady := func() {
		tray.OnReady(memory.EmptyStandbyList, memory.CheckAndCleanStandbyList)
	}

	systray.Run(onReady, onExit)
}

func onExit() {
	runtime.GC()
}
