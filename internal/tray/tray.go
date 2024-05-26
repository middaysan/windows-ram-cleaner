// internal/tray/tray.go

// Description: This file contains the tray package which is responsible for creating the system tray icon and handling the tray menu items.
package tray

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	windowsapi "windows-ram-cleaner/internal/windows_api"

	"github.com/getlantern/systray"
)

// Embed the icon using go:embed
//
//go:embed assets/icon.ico
var iconData []byte

const (
	PercentThreshold = 65
)

type TrayInfoStruct struct {
	LastSTDCleanup time.Time
	LastRAMCleanup time.Time
	ErrorStr string
}

var TrayInfo = TrayInfoStruct{
	LastSTDCleanup: time.Time{},
	LastRAMCleanup: time.Time{},
	ErrorStr: "",
}

// OnReady initializes the system tray icon and menu items.
// It sets the icon, title, and adds menu items for cleaning the standby list, cleaning the RAM, and quitting the application.
// It also starts a goroutine to handle menu item clicks.
func OnReady() {
	systray.SetIcon(iconData) // Use the embedded icon
	systray.SetTitle("Memory Cleaner")

	mSTDClean := systray.AddMenuItem("Clean Standby List", "Clean the standby list")
	mRAMClean := systray.AddMenuItem("Clean RAM", "Clean the RAM")
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	go handleMenuClicks(mSTDClean, mRAMClean, mQuit)
}

// UpdateTooltip updates the tooltip text of the system tray icon.
// It retrieves the standby list and free RAM size using the windowsapi package,
// and formats the tooltip string with the obtained values and the last cleanup timestamps.
// The formatted tooltip string is then set as the tooltip for the system tray icon.
func UpdateTooltip() {
	standbyList, freeRAM, err := windowsapi.GetStanbyListAndFreeRAMSize()
	if err != nil {
		TrayInfo.ErrorStr = err.Error()
	}

	tooltipStr := fmt.Sprintf(
		"SBL: %d MB\nFreeRAM: %d MB\nSTBcln: %s\nRAMcln: %s", 
		standbyList/(1024*1024), 
		freeRAM/(1024*1024),
		TrayInfo.LastSTDCleanup.Format("15:04:05"),
		TrayInfo.LastRAMCleanup.Format("15:04:05"),
	)

	systray.SetTooltip(tooltipStr)
}

func handleMenuClicks(mSTDClean, mRAMClean, mQuit *systray.MenuItem) {
	for {
		select {
		case <-mRAMClean.ClickedCh:
			if err := windowsapi.CleanRAM(); err != nil {
				TrayInfo.ErrorStr = err.Error()
				UpdateTooltip()
			} else {
				TrayInfo.LastRAMCleanup = time.Now()
				UpdateTooltip()
			}
		case <-mSTDClean.ClickedCh:
			if err := windowsapi.CleanStandbyList(); err != nil {
				TrayInfo.ErrorStr = err.Error()
				UpdateTooltip()
			} else {
				TrayInfo.LastSTDCleanup = time.Now()
				UpdateTooltip()
			}
		case <-mQuit.ClickedCh:
			systray.Quit()
			os.Exit(0)
		}
	}
}
