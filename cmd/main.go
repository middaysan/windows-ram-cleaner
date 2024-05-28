// cmd/main.go

package main

import (
	"windows-ram-cleaner/internal/tray"
	"windows-ram-cleaner/internal/windows_api"
	"runtime"
	"time"

	_ "github.com/josephspurrier/goversioninfo"

	"github.com/getlantern/systray"
)

var stopChan = make(chan struct{})
const (
	percentThreshold = 65
	autoCleanupCooldown = 5 * time.Minute
)

//go:generate goversioninfo -icon=exe_icon.ico -manifest=app.manifest
func main() {
	// Request admin rights if not already granted
	if !windowsapi.IsRunAsAdmin() {
		windowsapi.RunAsAdmin()
		return
	}

	go autoCleanStandbyList(stopChan)
	go autoUpdateTooltip(stopChan)

	systray.Run(tray.OnReady, onExit)
}

func onExit() {
	close(stopChan)
	runtime.GC()
}

// autoCleanStandbyList periodically cleans the standby list to free up RAM.
// It runs in a loop until the stopChan is closed.
// The function checks the percentage of standby list usage against the threshold.
// If the percentage is above the threshold and enough time has passed since the last cleanup,
// it calls the CleanStandbyList function to clean the standby list.
// It also updates the tooltip and sleeps for 1 minute before the next iteration.
//
// Parameters:
// - stopChan: A channel used to stop the function when closed.
//
// Note: The function assumes the availability of the windowsapi package.
func autoCleanStandbyList(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		default:
			time.Sleep(autoCleanupCooldown)
			standbyList, freeRAM, _ := windowsapi.GetStanbyListAndFreeRAMSize()
			percent := (standbyList * 100) / freeRAM

			if (percent > percentThreshold) {
				windowsapi.CleanStandbyList()
			}
		}
	}
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
