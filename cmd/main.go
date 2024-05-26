// cmd/main.go

package main

import (
	"clean-standby-list/internal/tray"
	"clean-standby-list/internal/windows_api"
	"runtime"
	"time"

	_ "github.com/josephspurrier/goversioninfo"

	"github.com/getlantern/systray"
)

var lastCleanup time.Time
var stopChan = make(chan struct{})
const (
	percentThreshold = 65
)

//go:generate goversioninfo -icon=exe_icon.ico -manifest=app.manifest
func main() {
	// Request admin rights if not already granted
	if !windowsapi.IsRunAsAdmin() {
		windowsapi.RunAsAdmin()
		return
	}

	go autoCleanStandbyList(stopChan)

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
			standbyList, freeRAM, _ := windowsapi.GetStanbyListAndFreeRAMSize()
			percent := (standbyList * 100) / freeRAM

			// 5 minutes cooldown
			if (percent > percentThreshold) && (time.Since(lastCleanup) > 5*time.Minute) {
				windowsapi.CleanStandbyList()
				lastCleanup = time.Now()
			}

			tray.UpdateTooltip()
			time.Sleep(5 * time.Second)
		}
	}
}
