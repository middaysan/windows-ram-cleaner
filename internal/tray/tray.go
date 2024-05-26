// internal/tray/tray.go

// Description: This file contains the tray package which is responsible for creating the system tray icon and handling the tray menu items.
package tray

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	windowsapi "windows-ram-cleaner/internal/windows_api"
	"windows-ram-cleaner/internal/win_startup"
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

type StartupManagement struct {
	AddToStartup *systray.MenuItem
	RemoveFromStartup *systray.MenuItem
}

type TrayMenuItems struct { 
	MSTDClean *systray.MenuItem
	MRAMClean *systray.MenuItem
	StartupManagement *StartupManagement
	MQuit *systray.MenuItem
}

var MenuItems = TrayMenuItems{}

// OnReady initializes the system tray icon and menu items.
// It sets the icon, title, and adds menu items for cleaning the standby list, cleaning the RAM, and quitting the application.
// It also starts a goroutine to handle menu item clicks.
func OnReady() {
	systray.SetIcon(iconData) // Use the embedded icon
	systray.SetTitle("Memory Cleaner")

	MenuItems.MSTDClean = systray.AddMenuItem("Clean Standby List", "Clean the standby list")
	MenuItems.MRAMClean = systray.AddMenuItem("Clean RAM", "Clean the RAM")
	MenuItems.StartupManagement = &StartupManagement{
		AddToStartup: systray.AddMenuItem("Add to Startup", "Add the application to startup"),
		RemoveFromStartup: systray.AddMenuItem("Remove from Startup", "Remove the application from startup"),
	}

	ex, err := winstartup.CheckStartupTask()
	if err != nil {
		MenuItems.StartupManagement.AddToStartup.Disable()
		MenuItems.StartupManagement.RemoveFromStartup.Disable()
		windowsapi.ShowError(
			fmt.Sprintf("Can't check startup aplication status, err: %s", err.Error()) , 
			"Error checking startup task",
		)
	} else {
		if ex {
			MenuItems.StartupManagement.AddToStartup.Disable()
		} else {
			MenuItems.StartupManagement.RemoveFromStartup.Disable()
		}
	}

	MenuItems.MQuit = systray.AddMenuItem("Quit", "Exit the application")

	go handleMenuClicks(&MenuItems)
}

// UpdateTooltip updates the tooltip text of the system tray icon.
// It retrieves the standby list and free RAM size using the windowsapi package,
// and formats the tooltip string with the obtained values and the last cleanup timestamps.
// The formatted tooltip string is then set as the tooltip for the system tray icon.
func UpdateTooltip() {
	standbyList, freeRAM, err := windowsapi.GetStanbyListAndFreeRAMSize()
	if err != nil {
		windowsapi.ShowError(
			fmt.Sprintf("Can't get standby list and free RAM size, err: %s", err.Error()) ,
			"Error getting standby list and free RAM size",
		)
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

func handleMenuClicks(TrayMenuItems *TrayMenuItems) {
	for {
		select {
		case <-TrayMenuItems.MRAMClean.ClickedCh:
			if err := windowsapi.CleanRAM(); err != nil {
				TrayInfo.ErrorStr = err.Error()
				UpdateTooltip()
			} else {
				TrayInfo.LastRAMCleanup = time.Now()
				UpdateTooltip()
			}
		case <-TrayMenuItems.MSTDClean.ClickedCh:
			if err := windowsapi.CleanStandbyList(); err != nil {
				TrayInfo.ErrorStr = err.Error()
				UpdateTooltip()
			} else {
				TrayInfo.LastSTDCleanup = time.Now()
				UpdateTooltip()
			}
		case <-TrayMenuItems.StartupManagement.AddToStartup.ClickedCh:
			if err := winstartup.CreateStartupTask(); err == nil {
				TrayMenuItems.StartupManagement.AddToStartup.Disable()
				TrayMenuItems.StartupManagement.RemoveFromStartup.Enable()
			} else {
				windowsapi.ShowError(
					fmt.Sprintf("Can't create startup task, err: %s", err.Error()) , 
					"Error creating startup task",
				)
			}
		case <-TrayMenuItems.StartupManagement.RemoveFromStartup.ClickedCh:
			if err := winstartup.DeleteStartupTask(); err == nil {
				TrayMenuItems.StartupManagement.RemoveFromStartup.Disable()
				TrayMenuItems.StartupManagement.AddToStartup.Enable()
			} else {
				windowsapi.ShowError(
					fmt.Sprintf("Can't delete startup task, err: %s", err.Error()) ,
					"Error deleting startup task",
				)
			}
		case <-TrayMenuItems.MQuit.ClickedCh:
			systray.Quit()
			os.Exit(0)
		}
	}
}
